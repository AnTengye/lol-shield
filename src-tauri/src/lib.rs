mod platform;
mod sidecar;

use futures_util::StreamExt;
use sidecar::{Controller, Snapshot};
use std::time::Duration;
use tauri::{Emitter, Manager};
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::Message;

const SIDECAR_WS_URL: &str = "ws://127.0.0.1:9365/ws";

fn start_status_bridge(app: tauri::AppHandle) {
    tauri::async_runtime::spawn(async move {
        loop {
            if !app.state::<Controller>().ready() {
                let _ = app.emit("shield-transport", false);
                tokio::time::sleep(Duration::from_secs(1)).await;
                continue;
            }
            let revision = app.state::<Controller>().snapshot().revision;
            if let Ok((mut stream, _)) = connect_async(SIDECAR_WS_URL).await {
                let _ = app.emit("shield-transport", true);
                loop {
                    let state = app.state::<Controller>().snapshot();
                    if state.phase != sidecar::Phase::Ready || state.revision != revision {
                        break;
                    }
                    let message =
                        match tokio::time::timeout(Duration::from_secs(1), stream.next()).await {
                            Ok(message) => message,
                            Err(_) => continue,
                        };
                    match message {
                        Some(Ok(Message::Text(text))) => {
                            if let Ok(payload) = serde_json::from_str::<serde_json::Value>(&text) {
                                let _ = app.emit(
                                    "shield-status",
                                    serde_json::json!({
                                        "revision": revision,
                                        "status": payload,
                                    }),
                                );
                            }
                        }
                        None | Some(Ok(Message::Close(_))) | Some(Err(_)) => break,
                        _ => {}
                    }
                }
            }

            let _ = app.emit("shield-transport", false);
            tokio::time::sleep(Duration::from_secs(3)).await;
        }
    });
}

#[tauri::command]
fn sidecar_status(app: tauri::AppHandle) -> Snapshot {
    app.state::<Controller>().snapshot()
}

#[tauri::command]
async fn prepare_update(app: tauri::AppHandle) -> Result<(), String> {
    let controller = app.state::<Controller>().inner().clone();
    tauri::async_runtime::spawn_blocking(move || controller.stop_for_update())
        .await
        .map_err(|e| e.to_string())?
}

#[tauri::command]
fn resume_sidecar(app: tauri::AppHandle) -> Result<(), String> {
    app.state::<Controller>().resume()
}

#[tauri::command]
fn open_startup_logs(app: tauri::AppHandle) -> Result<(), String> {
    let dir = app
        .path()
        .app_local_data_dir()
        .map_err(|e| e.to_string())?
        .join("logs");
    platform::open_logs(&dir).map_err(|e| e.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let instance = platform::SingleInstance::acquire();
    if matches!(instance, Ok(None)) {
        return;
    }
    let instance_error = instance
        .as_ref()
        .err()
        .map(|e| format!("无法取得单实例锁: {e}"));
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_process::init())
        .invoke_handler(tauri::generate_handler![
            prepare_update,
            resume_sidecar,
            sidecar_status,
            open_startup_logs
        ])
        .setup(move |app| {
            app.manage(Controller::new(app.handle().clone(), instance_error));
            start_status_bridge(app.handle().clone());
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            if matches!(event, tauri::RunEvent::Exit) {
                app.state::<Controller>().shutdown();
            }
        });
}

// 启动握手、会话退出与进程参数的回归测试位于 sidecar 和 platform 模块。
