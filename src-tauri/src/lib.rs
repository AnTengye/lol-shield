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
                let _ = app.emit("shield-transport", true);
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

            let _ = app.emit("shield-transport", false);
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

fn sidecar_running(app: &tauri::AppHandle) -> bool {
    let state = app.state::<SidecarState>();
    let guard = state.0.lock().expect("sidecar state poisoned");
    guard.is_some()
}

fn spawn_sidecar(app: &tauri::AppHandle) -> Result<(), Box<dyn std::error::Error>> {
    if sidecar_running(app) {
        return Ok(());
    }

    let data_dir = app.path().app_local_data_dir()?;
    std::fs::create_dir_all(&data_dir)?;
    let sidecar = app
        .shell()
        .sidecar("lol-shield")
        .expect("failed to create lol-shield sidecar command")
        .arg("--tauri-sidecar")
        .arg("--data-dir")
        .arg(data_dir.to_string_lossy().as_ref());

    let (_rx, child) = sidecar.spawn().expect("failed to spawn lol-shield sidecar");

    let state = app.state::<SidecarState>();
    *state.0.lock().expect("sidecar state poisoned") = Some(child);

    Ok(())
}

/// 安装更新前释放 sidecar 占用的文件句柄
///
/// Windows 下安装程序会替换安装目录里的 sidecar 可执行文件，
/// 更新插件在启动安装程序后会直接结束进程，来不及走正常退出流程，
/// 因此必须先把 sidecar 停掉。
#[tauri::command]
fn prepare_update(app: tauri::AppHandle) {
    stop_sidecar(&app);
}

/// 更新失败或用户放弃安装后恢复本地服务
#[tauri::command]
fn resume_sidecar(app: tauri::AppHandle) -> Result<(), String> {
    spawn_sidecar(&app).map_err(|err| err.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_process::init())
        .invoke_handler(tauri::generate_handler![prepare_update, resume_sidecar])
        .manage(SidecarState(Mutex::new(None)))
        .setup(|app| {
            spawn_sidecar(app.handle())?;

            start_status_bridge(app.handle().clone());

            Ok(())
        })
        .on_window_event(|window, event| {
            if matches!(event, tauri::WindowEvent::Destroyed) {
                stop_sidecar(window.app_handle());
            }
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            if matches!(event, tauri::RunEvent::Exit) {
                stop_sidecar(app);
            }
        });
}
