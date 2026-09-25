use std::io::{Read, Write};
use std::net::{SocketAddr, TcpStream};
use std::sync::Mutex;
use std::time::{Duration, Instant};

use futures_util::StreamExt;
use tauri::{Emitter, Manager};
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::Message;

const SIDECAR_WS_URL: &str = "ws://127.0.0.1:9365/ws";
const SIDECAR_PORT: u16 = 9365;
/// 安装器启动后等待 sidecar 释放文件占用的上限；超时则交给 NSIS 预安装钩子兜底。
const SIDECAR_EXIT_TIMEOUT: Duration = Duration::from_secs(10);

struct SidecarState(Mutex<Option<CommandChild>>);

fn sidecar_addr() -> SocketAddr {
    SocketAddr::from(([127, 0, 0, 1], SIDECAR_PORT))
}

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

/// sidecar 在 Windows 上会经 ShellExecute "runas" 自我提权重启：
/// 我们句柄里的中间进程随即退出，真正提供服务的是提权后的独立进程。
/// 因此不能只 kill 句柄，必须请求它经 /v1/shutdown 自行退出。
fn request_sidecar_shutdown_at(addr: SocketAddr) {
    let Ok(mut stream) = TcpStream::connect_timeout(&addr, Duration::from_secs(1)) else {
        return; // sidecar 未在运行
    };
    let _ = stream.set_write_timeout(Some(Duration::from_secs(2)));
    let _ = stream.set_read_timeout(Some(Duration::from_secs(2)));
    let request = format!(
        "POST /v1/shutdown HTTP/1.1\r\nHost: 127.0.0.1:{}\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
        addr.port()
    );
    if stream.write_all(request.as_bytes()).is_err() {
        return;
    }
    // 读到任意响应字节即说明服务端已受理退出请求，无需解析响应体。
    let mut buf = [0u8; 128];
    let _ = stream.read(&mut buf);
}

fn request_sidecar_shutdown() {
    request_sidecar_shutdown_at(sidecar_addr());
}

/// HTTP 端口仍在监听即认为提权 sidecar 存活（中间进程退出不影响此判定）。
fn http_alive(addr: SocketAddr) -> bool {
    TcpStream::connect_timeout(&addr, Duration::from_millis(300)).is_ok()
}

fn sidecar_http_alive() -> bool {
    http_alive(sidecar_addr())
}

/// 等待目标端口不再接受连接（进程已退出、文件句柄已释放）。
fn wait_for_http_exit(addr: SocketAddr, timeout: Duration) {
    let deadline = Instant::now() + timeout;
    while Instant::now() < deadline {
        if !http_alive(addr) {
            return;
        }
        std::thread::sleep(Duration::from_millis(200));
    }
}

fn wait_for_sidecar_exit(timeout: Duration) {
    wait_for_http_exit(sidecar_addr(), timeout);
}

fn stop_sidecar(app: &tauri::AppHandle) {
    // 先请求提权 sidecar 自行退出，再结束本地句柄（若仍存活）。
    request_sidecar_shutdown();
    let state = app.state::<SidecarState>();
    let child = state.0.lock().expect("sidecar state poisoned").take();
    if let Some(child) = child {
        let _ = child.kill();
    }
}

fn spawn_sidecar(app: &tauri::AppHandle) -> Result<(), Box<dyn std::error::Error>> {
    // 提权后的 sidecar 可能已在运行（例如上一会话残留），直接复用，避免重复弹 UAC。
    if sidecar_http_alive() {
        return Ok(());
    }
    let state = app.state::<SidecarState>();
    {
        let guard = state.0.lock().expect("sidecar state poisoned");
        // 句柄仍在说明本次会话已拉起过（提权确认窗口期内），不要重复启动。
        if guard.is_some() {
            return Ok(());
        }
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

    *state.0.lock().expect("sidecar state poisoned") = Some(child);

    Ok(())
}

/// 安装更新前释放 sidecar 占用的文件句柄。
///
/// Windows 下安装程序会替换安装目录里的 sidecar 可执行文件，
/// 更新插件在启动安装程序后会直接结束进程，来不及走正常退出流程。
/// 提权 sidecar 无法被进程外直接结束，只能请求其自行退出，
/// 并等待端口释放后再交给安装器；万一超时，NSIS 预安装钩子还会兜底。
#[tauri::command]
fn prepare_update(app: tauri::AppHandle) {
    stop_sidecar(&app);
    wait_for_sidecar_exit(SIDECAR_EXIT_TIMEOUT);
}

/// 更新失败或用户放弃安装后恢复本地服务。
/// 无条件重建：旧句柄对应的中间进程可能早已退出，不能作为存活依据。
#[tauri::command]
fn resume_sidecar(app: tauri::AppHandle) -> Result<(), String> {
    {
        let state = app.state::<SidecarState>();
        let _ = state
            .0
            .lock()
            .expect("sidecar state poisoned")
            .take();
    }
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

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::TcpListener;

    #[test]
    fn shutdown_request_posts_expected_http_request() {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        let server = std::thread::spawn(move || {
            let (mut stream, _) = listener.accept().unwrap();
            let mut buf = [0u8; 1024];
            let n = stream.read(&mut buf).unwrap();
            let request = String::from_utf8_lossy(&buf[..n]).into_owned();
            let _ = stream.write_all(
                b"HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
            );
            request
        });

        request_sidecar_shutdown_at(addr);

        let request = server.join().unwrap();
        assert!(request.starts_with("POST /v1/shutdown HTTP/1.1"));
        assert!(request.contains("Content-Length: 0"));
    }

    #[test]
    fn shutdown_request_tolerates_offline_sidecar() {
        // 绑定后立即释放，拿到一个几乎必然空闲的端口
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        drop(listener);

        // sidecar 不在运行时应静默返回，不 panic、不长时间阻塞
        request_sidecar_shutdown_at(addr);
    }

    #[test]
    fn wait_for_http_exit_returns_before_timeout_once_port_closes() {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        std::thread::spawn(move || {
            std::thread::sleep(Duration::from_millis(300));
            drop(listener);
        });

        let start = Instant::now();
        wait_for_http_exit(addr, Duration::from_secs(5));
        assert!(
            start.elapsed() < Duration::from_secs(5),
            "端口关闭后应提前返回而不是等到超时"
        );
    }
}
