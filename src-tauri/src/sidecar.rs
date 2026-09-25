use crate::platform::ManagedProcess;
use serde::{Deserialize, Serialize};
use std::{
    fs::OpenOptions,
    io::{Read, Write},
    net::{SocketAddr, TcpStream},
    path::PathBuf,
    sync::{mpsc, Arc, Mutex},
    time::{Duration, Instant, SystemTime, UNIX_EPOCH},
};
use tauri::{Emitter, Manager};
use tauri_plugin_shell::ShellExt;

pub const PORT: u16 = 9365;
const READY_TIMEOUT: Duration = Duration::from_secs(20);
const EXIT_TIMEOUT: Duration = Duration::from_secs(6);

#[derive(Clone, Copy, Debug, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Phase {
    Starting,
    Authorizing,
    Ready,
    Failed,
    Stopping,
    Updating,
}

#[derive(Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct Snapshot {
    pub phase: Phase,
    pub message: String,
    pub pid: Option<u32>,
    pub log_path: String,
    pub revision: u64,
}

#[derive(Default)]
struct Readiness {
    deadline: Option<Instant>,
    misses: u8,
}

#[derive(Debug, PartialEq)]
enum HealthTransition {
    Pending,
    Ready,
    Failed(&'static str),
}

impl Readiness {
    fn started(now: Instant) -> Self {
        Self {
            deadline: Some(now + READY_TIMEOUT),
            misses: 0,
        }
    }

    fn observe(&mut self, phase: Phase, healthy: bool, now: Instant) -> HealthTransition {
        if !matches!(phase, Phase::Starting | Phase::Ready) {
            return HealthTransition::Pending;
        }
        if healthy {
            self.misses = 0;
            self.deadline = None;
            return if phase == Phase::Ready {
                HealthTransition::Pending
            } else {
                HealthTransition::Ready
            };
        }
        if phase == Phase::Ready {
            self.misses = self.misses.saturating_add(1);
            if self.misses >= 3 {
                return HealthTransition::Failed("本地服务连续无响应，请查看日志并重试");
            }
        } else if self.deadline.is_some_and(|deadline| now >= deadline) {
            return HealthTransition::Failed("后台进程启动后 20 秒仍未就绪，请查看启动日志并重试");
        }
        HealthTransition::Pending
    }
}

type Reply = mpsc::Sender<Result<(), String>>;
enum Request {
    Resume,
    Stop(Reply),
    Exit(Reply),
}

#[derive(Clone)]
pub struct Controller {
    tx: mpsc::Sender<Request>,
    snapshot: Arc<Mutex<Snapshot>>,
}
impl Controller {
    pub fn new(app: tauri::AppHandle, startup_error: Option<String>) -> Self {
        let dir = match startup_error {
            Some(error) => Err(error),
            None => app.path().app_local_data_dir().map_err(|e| e.to_string()),
        };
        let snapshot = Arc::new(Mutex::new(Snapshot {
            phase: Phase::Starting,
            message: "正在准备本地服务".into(),
            pid: None,
            log_path: dir
                .as_ref()
                .map(|d| d.join("logs").display().to_string())
                .unwrap_or_default(),
            revision: 0,
        }));
        let (tx, rx) = mpsc::channel();
        let controller = Self {
            tx,
            snapshot: snapshot.clone(),
        };
        std::thread::spawn(move || {
            let mut worker = Worker {
                app,
                snapshot,
                dir,
                process: None,
                session: String::new(),
                readiness: Readiness::default(),
            };
            worker.start();
            loop {
                match rx.recv_timeout(Duration::from_millis(500)) {
                    Ok(Request::Resume) => worker.start(),
                    Ok(Request::Stop(reply)) => {
                        worker.set(Phase::Stopping, "正在停止本地服务");
                        let result = worker.stop();
                        match &result {
                            Ok(()) => worker.set(Phase::Updating, "本地服务已停止，等待安装更新"),
                            Err(error) => worker.set(Phase::Failed, error),
                        }
                        let _ = reply.send(result);
                    }
                    Ok(Request::Exit(reply)) => {
                        worker.set(Phase::Stopping, "正在退出本地服务");
                        let result = worker.stop();
                        let _ = reply.send(result);
                        break;
                    }
                    Err(mpsc::RecvTimeoutError::Disconnected) => {
                        let _ = worker.stop();
                        break;
                    }
                    Err(mpsc::RecvTimeoutError::Timeout) => worker.poll(),
                }
            }
        });
        controller
    }
    pub fn snapshot(&self) -> Snapshot {
        self.snapshot.lock().unwrap().clone()
    }
    pub fn ready(&self) -> bool {
        self.snapshot().phase == Phase::Ready
    }
    pub fn resume(&self) -> Result<(), String> {
        self.tx
            .send(Request::Resume)
            .map_err(|_| "本地服务管理线程已退出，请重启应用".into())
    }
    pub fn stop_for_update(&self) -> Result<(), String> {
        let (tx, rx) = mpsc::channel();
        self.tx
            .send(Request::Stop(tx))
            .map_err(|_| "本地服务管理线程已退出".to_string())?;
        rx.recv_timeout(Duration::from_secs(12)).map_err(|_| {
            "停止本地服务超时；如有授权窗口，请先完成或取消授权，再重试更新".to_string()
        })?
    }
    pub fn shutdown(&self) {
        let (tx, rx) = mpsc::channel();
        let _ = self.tx.send(Request::Exit(tx));
        let _ = rx.recv_timeout(Duration::from_secs(10));
    }
}

struct Worker {
    app: tauri::AppHandle,
    snapshot: Arc<Mutex<Snapshot>>,
    dir: Result<PathBuf, String>,
    process: Option<ManagedProcess>,
    session: String,
    readiness: Readiness,
}
impl Worker {
    fn set(&self, phase: Phase, message: impl Into<String>) {
        let next = {
            let mut state = self.snapshot.lock().unwrap();
            state.phase = phase;
            state.message = message.into();
            state.pid = self.process.as_ref().map(ManagedProcess::id);
            state.revision += 1;
            state.clone()
        };
        if let Ok(dir) = &self.dir {
            if let Ok(mut log) = OpenOptions::new()
                .create(true)
                .append(true)
                .open(dir.join("logs/desktop-startup.log"))
            {
                let seconds = SystemTime::now()
                    .duration_since(UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs();
                let _ = writeln!(
                    log,
                    "{seconds} shell={} version={} phase={phase:?} pid={:?} {}",
                    std::process::id(),
                    self.app.package_info().version,
                    next.pid,
                    next.message
                );
            }
        }
        let _ = self.app.emit("shield-sidecar", next);
    }
    fn start(&mut self) {
        // 重复恢复请求不重启正在启动或已就绪的服务，也不会重复弹 UAC。
        if self.process.is_some() && self.snapshot.lock().unwrap().phase != Phase::Failed {
            return;
        }
        if let Err(error) = self.stop() {
            self.set(Phase::Failed, error);
            return;
        }
        self.set(Phase::Starting, "正在准备本地服务");
        if let Err(error) = self.launch() {
            self.set(Phase::Failed, error);
        }
    }
    fn launch(&mut self) -> Result<(), String> {
        let dir = self.dir.clone()?;
        std::fs::create_dir_all(dir.join("logs"))
            .map_err(|e| format!("无法创建日志目录 {}: {e}", dir.display()))?;
        // 启动前就落盘，不依赖 Go 是否进入 main。
        OpenOptions::new()
            .create(true)
            .append(true)
            .open(dir.join("logs/desktop-startup.log"))
            .map_err(|e| format!("无法写入桌面启动日志: {e}"))?;
        self.set(Phase::Starting, format!("用户目录：{}", dir.display()));
        if TcpStream::connect_timeout(&addr(), Duration::from_millis(300)).is_ok() {
            return Err("9365 端口已被其他进程占用；请关闭旧版本或占用该端口的程序后重试。不会复用或关闭未知服务。".into());
        }
        let command: std::process::Command = self
            .app
            .shell()
            .sidecar("lol-shield")
            .map_err(|e| format!("找不到后台程序: {e}"))?
            .into();
        let exe = PathBuf::from(command.get_program());
        self.session = uuid::Uuid::new_v4().to_string();
        let args = vec![
            "--tauri-sidecar".into(),
            "--data-dir".into(),
            dir.display().to_string(),
            "-c".into(),
            dir.join("config.yaml").display().to_string(),
            "--desktop-session".into(),
            self.session.clone(),
            "--parent-pid".into(),
            std::process::id().to_string(),
        ];
        self.set(
            if cfg!(windows) {
                Phase::Authorizing
            } else {
                Phase::Starting
            },
            "正在启动本地服务；如出现 Windows 授权窗口，请允许",
        );
        self.process = Some(
            ManagedProcess::spawn(&exe, &args, &dir, &dir.join("logs/process-output.log"))
                .map_err(|e| {
                    if e.raw_os_error() == Some(1223) {
                        "已取消管理员授权，请点击重试并允许授权".to_string()
                    } else {
                        format!("Windows 后台启动失败: {e}")
                    }
                })?,
        );
        self.readiness = Readiness::started(Instant::now());
        self.set(Phase::Starting, "后台进程已创建，正在等待服务就绪");
        Ok(())
    }
    fn poll(&mut self) {
        let phase = self.snapshot.lock().unwrap().phase;
        if !matches!(phase, Phase::Starting | Phase::Ready) {
            return;
        }
        let Some(process) = self.process.as_mut() else {
            return;
        };
        match process.try_wait() {
            Ok(Some(code)) => {
                self.process.take();
                self.readiness = Readiness::default();
                self.set(
                    Phase::Failed,
                    format!("本地服务已退出（退出码 {code}），请查看启动日志后重试"),
                );
                return;
            }
            Err(error) => {
                self.fail_and_stop(&format!("读取后台进程状态失败: {error}"));
                return;
            }
            Ok(None) => {}
        }
        let pid = self.process.as_ref().unwrap().id();
        let healthy = health_at(addr(), &self.session, pid);
        match self.readiness.observe(phase, healthy, Instant::now()) {
            HealthTransition::Pending => {}
            HealthTransition::Ready => self.set(Phase::Ready, "本地服务正常"),
            HealthTransition::Failed(reason) => self.fail_and_stop(reason),
        }
    }
    fn fail_and_stop(&mut self, reason: &str) {
        let message = match self.stop() {
            Ok(()) => reason.to_string(),
            Err(e) => format!("{reason}；{e}"),
        };
        self.set(Phase::Failed, message);
    }
    fn stop(&mut self) -> Result<(), String> {
        let Some(process) = self.process.as_mut() else {
            return Ok(());
        };
        if process.try_wait().map_err(|e| e.to_string())?.is_none() {
            // 会话校验只允许关闭本次启动的服务，旧窗口无法关闭新服务。
            let _ = request_at(addr(), "POST", "/v1/shutdown", &self.session);
            if !wait_exit(process, EXIT_TIMEOUT)? {
                process
                    .kill()
                    .map_err(|e| format!("后台未能退出，停止安装以避免文件占用: {e}"))?;
                if !wait_exit(process, Duration::from_secs(2))? {
                    return Err("后台进程仍未退出，已阻止更新安装".into());
                }
            }
        }
        self.process.take();
        self.readiness = Readiness::default();
        Ok(())
    }
}

fn addr() -> SocketAddr {
    SocketAddr::from(([127, 0, 0, 1], PORT))
}

fn wait_exit(process: &mut ManagedProcess, timeout: Duration) -> Result<bool, String> {
    let deadline = Instant::now() + timeout;
    loop {
        if process.try_wait().map_err(|e| e.to_string())?.is_some() {
            return Ok(true);
        }
        if Instant::now() >= deadline {
            return Ok(false);
        }
        std::thread::sleep(Duration::from_millis(50));
    }
}

#[derive(Deserialize)]
struct Health {
    service: String,
    protocol: u32,
    pid: u32,
}
fn health_at(addr: SocketAddr, session: &str, pid: u32) -> bool {
    request_at(addr, "GET", "/v1/health", session)
        .ok()
        .and_then(|body| serde_json::from_slice::<Health>(&body).ok())
        .is_some_and(|health| {
            health.service == "lol-shield" && health.protocol == 1 && health.pid == pid
        })
}

fn request_at(
    addr: SocketAddr,
    method: &str,
    path: &str,
    session: &str,
) -> Result<Vec<u8>, String> {
    let mut stream =
        TcpStream::connect_timeout(&addr, Duration::from_millis(300)).map_err(|e| e.to_string())?;
    stream
        .set_read_timeout(Some(Duration::from_millis(700)))
        .map_err(|e| e.to_string())?;
    stream
        .set_write_timeout(Some(Duration::from_millis(300)))
        .map_err(|e| e.to_string())?;
    let request = format!("{method} {path} HTTP/1.1\r\nHost: 127.0.0.1:{}\r\nX-Shield-Session: {session}\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", addr.port());
    stream
        .write_all(request.as_bytes())
        .map_err(|e| e.to_string())?;
    let mut response = Vec::new();
    stream
        .take(8192)
        .read_to_end(&mut response)
        .map_err(|e| e.to_string())?;
    parse_response(&response)
}
fn parse_response(response: &[u8]) -> Result<Vec<u8>, String> {
    let split = response
        .windows(4)
        .position(|bytes| bytes == b"\r\n\r\n")
        .ok_or("服务响应无效")?;
    let header = std::str::from_utf8(&response[..split]).map_err(|e| e.to_string())?;
    let status = header
        .lines()
        .next()
        .and_then(|line| line.split_whitespace().nth(1));
    if status != Some("200") {
        return Err("服务拒绝请求或身份不匹配".into());
    }
    Ok(response[split + 4..].to_vec())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::TcpListener;

    fn serve(body: &'static str, status: &str) -> (SocketAddr, std::thread::JoinHandle<String>) {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        let status = status.to_string();
        let server = std::thread::spawn(move || {
            let (mut stream, _) = listener.accept().unwrap();
            let mut buf = [0; 4096];
            let n = stream.read(&mut buf).unwrap();
            write!(
                stream,
                "HTTP/1.1 {status}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{body}",
                body.len()
            )
            .unwrap();
            String::from_utf8_lossy(&buf[..n]).into_owned()
        });
        (addr, server)
    }
    #[test]
    fn health_requires_service_protocol_and_owned_pid() {
        for body in [
            r#"{"service":"other","protocol":1,"pid":12}"#,
            r#"{"service":"lol-shield","protocol":2,"pid":12}"#,
            r#"{"service":"lol-shield","protocol":1,"pid":99}"#,
            "not json",
        ] {
            let (addr, server) = serve(body, "200 OK");
            assert!(!health_at(addr, "test-session", 12));
            server.join().unwrap();
        }
        let (addr, server) = serve(
            r#"{"service":"lol-shield","protocol":1,"pid":12}"#,
            "200 OK",
        );
        assert!(health_at(addr, "test-session", 12));
        assert!(server
            .join()
            .unwrap()
            .contains("X-Shield-Session: test-session"));
    }
    #[test]
    fn shutdown_sends_session_and_rejects_foreign_service() {
        let (addr, server) = serve("", "403 Forbidden");
        assert!(request_at(addr, "POST", "/v1/shutdown", "owner").is_err());
        let request = server.join().unwrap();
        assert!(request.starts_with("POST /v1/shutdown HTTP/1.1"));
        assert!(request.contains("X-Shield-Session: owner"));
    }
    #[test]
    fn startup_timeout_has_a_finite_boundary() {
        let start = Instant::now();
        let mut readiness = Readiness::started(start);
        assert_eq!(
            readiness.observe(
                Phase::Starting,
                false,
                start + READY_TIMEOUT - Duration::from_millis(1)
            ),
            HealthTransition::Pending
        );
        assert!(matches!(
            readiness.observe(Phase::Starting, false, start + READY_TIMEOUT),
            HealthTransition::Failed(_)
        ));
    }

    #[test]
    fn ready_requires_health_and_clears_startup_deadline() {
        let now = Instant::now();
        let mut readiness = Readiness::started(now);
        assert_eq!(
            readiness.observe(Phase::Starting, true, now),
            HealthTransition::Ready
        );
        assert!(readiness.deadline.is_none());
        assert_eq!(
            readiness.observe(Phase::Ready, true, now + READY_TIMEOUT),
            HealthTransition::Pending
        );
    }

    #[test]
    fn only_consecutive_health_failures_stop_a_ready_service() {
        let now = Instant::now();
        let mut readiness = Readiness::default();
        for _ in 0..2 {
            assert_eq!(
                readiness.observe(Phase::Ready, false, now),
                HealthTransition::Pending
            );
        }
        assert_eq!(
            readiness.observe(Phase::Ready, true, now),
            HealthTransition::Pending
        );
        for _ in 0..2 {
            assert_eq!(
                readiness.observe(Phase::Ready, false, now),
                HealthTransition::Pending
            );
        }
        assert!(matches!(
            readiness.observe(Phase::Ready, false, now),
            HealthTransition::Failed(_)
        ));
    }

    #[test]
    fn late_health_cannot_revive_failed_or_updating_service() {
        let now = Instant::now();
        let mut readiness = Readiness::started(now);
        for phase in [
            Phase::Authorizing,
            Phase::Failed,
            Phase::Stopping,
            Phase::Updating,
        ] {
            for healthy in [true, false] {
                assert_eq!(
                    readiness.observe(phase, healthy, now + READY_TIMEOUT),
                    HealthTransition::Pending
                );
            }
        }
    }

    #[test]
    fn malformed_http_is_not_ready() {
        assert!(parse_response(b"HTTP/1.1 200 OK").is_err());
        assert!(parse_response(b"HTTP/1.1 500 Error\r\n\r\n{}").is_err());
    }
}
