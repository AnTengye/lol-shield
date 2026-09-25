use std::{io, path::Path};

// Windows 命令行的引号与尾部反斜线必须成对转义，支持空格和中文目录。
#[cfg(any(windows, test))]
fn quote_windows_arg(value: &str) -> String {
    let mut result = String::from("\"");
    let mut slashes = 0;
    for ch in value.chars() {
        if ch == '\\' {
            slashes += 1;
            continue;
        }
        if ch == '"' {
            result.push_str(&"\\".repeat(slashes * 2 + 1));
        } else {
            result.push_str(&"\\".repeat(slashes));
        }
        slashes = 0;
        result.push(ch);
    }
    result.push_str(&"\\".repeat(slashes * 2));
    result.push('"');
    result
}

#[cfg(windows)]
mod native {
    use super::*;
    use std::{ffi::OsStr, os::windows::ffi::OsStrExt, ptr};
    use windows_sys::Win32::{
        Foundation::{
            CloseHandle, GetLastError, ERROR_ALREADY_EXISTS, HANDLE, WAIT_OBJECT_0, WAIT_TIMEOUT,
        },
        System::{
            Com::{CoInitializeEx, CoUninitialize, COINIT_APARTMENTTHREADED},
            Threading::{
                CreateMutexW, GetExitCodeProcess, GetProcessId, TerminateProcess,
                WaitForSingleObject,
            },
        },
        UI::{
            Shell::{
                ShellExecuteExW, ShellExecuteW, SEE_MASK_NOASYNC, SEE_MASK_NOCLOSEPROCESS,
                SHELLEXECUTEINFOW,
            },
            WindowsAndMessaging::{
                FindWindowW, SetForegroundWindow, ShowWindowAsync, SW_HIDE, SW_RESTORE,
                SW_SHOWNORMAL,
            },
        },
    };

    fn wide(value: impl AsRef<OsStr>) -> Vec<u16> {
        value.as_ref().encode_wide().chain(Some(0)).collect()
    }

    pub struct SingleInstance(HANDLE);
    impl SingleInstance {
        pub fn acquire() -> io::Result<Option<Self>> {
            Self::acquire_named("Local\\work.bigorange.lol-shield.desktop")
        }
        pub(super) fn acquire_named(name: &str) -> io::Result<Option<Self>> {
            let name = wide(name);
            let handle = unsafe { CreateMutexW(ptr::null(), 0, name.as_ptr()) };
            if handle.is_null() {
                return Err(io::Error::last_os_error());
            }
            if unsafe { GetLastError() } == ERROR_ALREADY_EXISTS {
                unsafe {
                    CloseHandle(handle);
                    let window = FindWindowW(ptr::null(), wide("LOL Shield").as_ptr());
                    if !window.is_null() {
                        ShowWindowAsync(window, SW_RESTORE);
                        SetForegroundWindow(window);
                    }
                }
                return Ok(None);
            }
            Ok(Some(Self(handle)))
        }
    }
    impl Drop for SingleInstance {
        fn drop(&mut self) {
            unsafe {
                CloseHandle(self.0);
            }
        }
    }

    pub struct ManagedProcess {
        handle: HANDLE,
        pid: u32,
    }
    impl ManagedProcess {
        pub fn spawn(exe: &Path, args: &[String], dir: &Path, _log: &Path) -> io::Result<Self> {
            let file = wide(exe);
            let cwd = wide(dir);
            let verb = wide("runas");
            let parameters = wide(
                args.iter()
                    .map(|arg| quote_windows_arg(arg))
                    .collect::<Vec<_>>()
                    .join(" "),
            );
            // 在监督线程内等待 UAC，窗口与状态查询不受阻塞。
            let com = unsafe { CoInitializeEx(ptr::null(), COINIT_APARTMENTTHREADED as u32) };
            if com < 0 {
                return Err(io::Error::other(format!(
                    "初始化提权环境失败: HRESULT {com:#x}"
                )));
            }
            let mut info = SHELLEXECUTEINFOW {
                cbSize: std::mem::size_of::<SHELLEXECUTEINFOW>() as u32,
                fMask: SEE_MASK_NOCLOSEPROCESS | SEE_MASK_NOASYNC,
                lpVerb: verb.as_ptr(),
                lpFile: file.as_ptr(),
                lpParameters: parameters.as_ptr(),
                lpDirectory: cwd.as_ptr(),
                nShow: SW_HIDE,
                ..Default::default()
            };
            let success = unsafe { ShellExecuteExW(&mut info) };
            let error = io::Error::last_os_error();
            unsafe {
                CoUninitialize();
            }
            if success == 0 {
                return Err(error);
            }
            if info.hProcess.is_null() {
                return Err(io::Error::other("提权启动未返回进程句柄"));
            }
            let pid = unsafe { GetProcessId(info.hProcess) };
            if pid == 0 {
                let error = io::Error::last_os_error();
                unsafe {
                    CloseHandle(info.hProcess);
                }
                return Err(error);
            }
            Ok(Self {
                handle: info.hProcess,
                pid,
            })
        }
        pub fn id(&self) -> u32 {
            self.pid
        }
        pub fn try_wait(&mut self) -> io::Result<Option<u32>> {
            match unsafe { WaitForSingleObject(self.handle, 0) } {
                WAIT_TIMEOUT => Ok(None),
                WAIT_OBJECT_0 => {
                    let mut code = 0;
                    if unsafe { GetExitCodeProcess(self.handle, &mut code) } == 0 {
                        return Err(io::Error::last_os_error());
                    }
                    Ok(Some(code))
                }
                _ => Err(io::Error::last_os_error()),
            }
        }
        pub fn kill(&mut self) -> io::Result<()> {
            if unsafe { TerminateProcess(self.handle, 1) } == 0 {
                return Err(io::Error::last_os_error());
            }
            Ok(())
        }
    }
    impl Drop for ManagedProcess {
        fn drop(&mut self) {
            unsafe {
                CloseHandle(self.handle);
            }
        }
    }

    pub fn open_logs(dir: &Path) -> io::Result<()> {
        let result = unsafe {
            ShellExecuteW(
                ptr::null_mut(),
                wide("open").as_ptr(),
                wide(dir).as_ptr(),
                ptr::null(),
                ptr::null(),
                SW_SHOWNORMAL,
            )
        };
        if result as isize <= 32 {
            return Err(io::Error::other("无法打开日志目录"));
        }
        Ok(())
    }
}

// 非 Windows 路径仅保留本地开发和测试能力，正式产品仍只面向 Windows。
#[cfg(not(windows))]
mod native {
    use super::*;
    use std::{
        fs::OpenOptions,
        process::{Child, Command, Stdio},
    };
    pub struct SingleInstance;
    impl SingleInstance {
        pub fn acquire() -> io::Result<Option<Self>> {
            Ok(Some(Self))
        }
    }
    pub struct ManagedProcess(Child);
    impl ManagedProcess {
        pub fn spawn(exe: &Path, args: &[String], dir: &Path, log: &Path) -> io::Result<Self> {
            let output = OpenOptions::new().create(true).append(true).open(log)?;
            Command::new(exe)
                .args(args)
                .current_dir(dir)
                .stdin(Stdio::null())
                .stdout(output.try_clone()?)
                .stderr(output)
                .spawn()
                .map(Self)
        }
        pub fn id(&self) -> u32 {
            self.0.id()
        }
        pub fn try_wait(&mut self) -> io::Result<Option<u32>> {
            self.0
                .try_wait()
                .map(|status| status.map(|s| s.code().unwrap_or(-1) as u32))
        }
        pub fn kill(&mut self) -> io::Result<()> {
            self.0.kill()
        }
    }
    pub fn open_logs(_dir: &Path) -> io::Result<()> {
        Err(io::Error::other("请按界面中的路径查看日志"))
    }
}

pub use native::*;

#[cfg(test)]
mod tests {
    use super::*;
    #[cfg(windows)]
    #[test]
    fn instance_lock_is_released_after_owner_exits() {
        let name = format!("Local\\shield-test-{}", uuid::Uuid::new_v4());
        let owner = native::SingleInstance::acquire_named(&name)
            .unwrap()
            .unwrap();
        assert!(native::SingleInstance::acquire_named(&name)
            .unwrap()
            .is_none());
        drop(owner);
        assert!(native::SingleInstance::acquire_named(&name)
            .unwrap()
            .is_some());
    }

    #[cfg(windows)]
    #[test]
    fn arguments_roundtrip_through_windows_parser() {
        use windows_sys::Win32::{Foundation::LocalFree, UI::Shell::CommandLineToArgvW};
        let args = [
            "sidecar.exe",
            "",
            "中文 用户",
            "C:\\path with space\\",
            "a\\\"b",
        ];
        let command = args
            .iter()
            .map(|arg| quote_windows_arg(arg))
            .collect::<Vec<_>>()
            .join(" ");
        let wide: Vec<u16> = command.encode_utf16().chain(Some(0)).collect();
        unsafe {
            let mut count = 0;
            let parsed = CommandLineToArgvW(wide.as_ptr(), &mut count);
            assert!(!parsed.is_null());
            let actual: Vec<String> = std::slice::from_raw_parts(parsed, count as usize)
                .iter()
                .map(|value| {
                    let mut len = 0;
                    while *value.add(len) != 0 {
                        len += 1;
                    }
                    String::from_utf16(std::slice::from_raw_parts(*value, len)).unwrap()
                })
                .collect();
            LocalFree(parsed.cast());
            assert_eq!(actual, args);
        }
    }

    #[test]
    fn windows_arguments_preserve_spaces_quotes_and_trailing_slashes() {
        assert_eq!(quote_windows_arg(""), "\"\"");
        assert_eq!(
            quote_windows_arg("中文路径 with space"),
            "\"中文路径 with space\""
        );
        assert_eq!(quote_windows_arg("a\"b"), "\"a\\\"b\"");
        assert_eq!(quote_windows_arg("C:\\data\\"), "\"C:\\data\\\\\"");
        assert_eq!(quote_windows_arg("a\\\"b"), "\"a\\\\\\\"b\"");
    }
}
