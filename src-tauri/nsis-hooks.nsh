; NSIS 安装钩子。
;
; 背景：sidecar（lol-shield.exe）在 Windows 上经 ShellExecute "runas" 自我提权，
; 以管理员身份独立运行，桌面壳无法在进程外结束它；Windows 不允许写入正被
; 运行中进程锁定的可执行文件，导致覆盖安装报
; "Error opening file for writing: ...lol-shield.exe"。
;
; 桌面壳的 prepare_update 已请求 sidecar 优雅退出并等待端口释放，
; 此钩子作为兜底：无论优雅退出失败还是旧版本/手动安装场景，
; 在复制文件前强制结束残留的 sidecar 进程（安装器自身已是管理员权限）。

!macro NSIS_HOOK_PREINSTALL
  ; /IM 按映像名结束进程；/T 连同子进程。进程不存在时 taskkill 返回非零，属预期，忽略即可。
  nsExec::Exec 'taskkill /F /IM "lol-shield.exe" /T'
  ; TerminateProcess 发出后句柄释放有微小延迟，稍候再进入文件复制阶段。
  Sleep 500
!macroend
