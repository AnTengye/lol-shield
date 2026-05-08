use std::sync::Mutex;
use std::time::Duration;

use futures_util::StreamExt;
use tauri::{Emitter, Manager};
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::Message;

const SIDECAR_WS_URL: &str = "ws://127.0.0.1:9365/ws";

struct SidecarState(Mutex<Option<CommandChild>>);

fn start_status_bridge(app: tauri::AppHandle) {
    tauri::async_runtime::spawn(async move {
        loop {
            if let Ok((mut stream, _)) = connect_async(SIDECAR_WS_URL).await {
                while let Some(message) = stream.next().await {
                    match message {
                        Ok(Message::Text(text)) => {
                            if let Ok(payload) = serde_json::from_str::<serde_json::Value>(&text) {
                                let _ = app.emit("shield-status", payload);
                            }
                        }
                        Ok(Message::Close(_)) | Err(_) => break,
                        _ => {}
                    }
                }
            }

            tokio::time::sleep(Duration::from_secs(3)).await;
        }
    });
}

fn stop_sidecar(app: &tauri::AppHandle) {
    let state = app.state::<SidecarState>();
    let child = state.0.lock().expect("sidecar state poisoned").take();
    if let Some(child) = child {
        let _ = child.kill();
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(SidecarState(Mutex::new(None)))
        .setup(|app| {
            let sidecar = app
                .shell()
                .sidecar("lol-shield")
                .expect("failed to create lol-shield sidecar command")
                .arg("--tauri-sidecar");

            let (_rx, child) = sidecar.spawn().expect("failed to spawn lol-shield sidecar");

            let state = app.state::<SidecarState>();
            *state.0.lock().expect("sidecar state poisoned") = Some(child);

            start_status_bridge(app.handle().clone());

            Ok(())
        })
        .on_window_event(|window, event| {
            if matches!(event, tauri::WindowEvent::CloseRequested { .. }) {
                stop_sidecar(window.app_handle());
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
