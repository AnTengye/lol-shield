# Tauri Sidecar Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the existing Vue UI as a Tauri desktop app while keeping the current Go Shield service as a managed sidecar.

**Architecture:** Tauri owns the desktop window, bundling, installer, sidecar lifecycle, and frontend event delivery. The existing Go service keeps all LCU, Windows process, config, and HTTP API behavior during phase 1, but Go's update check is removed. Vue continues using HTTP for request/response APIs, while live status moves from direct browser WebSocket usage to Tauri events bridged by Rust.

**Tech Stack:** Vue 3, Vite, Go, Gin, Gorilla WebSocket, Tauri v2, Rust, pnpm/npm, Windows sidecar packaging.

---

## Scope

Phase 1 intentionally does not rewrite LCU logic in Rust. It keeps the Go backend as the source of truth for:

- LCU process token discovery.
- LCU HTTPS API calls.
- LCU WebSocket event monitoring.
- Auto accept, auto pick, and auto ban.
- `/v1` and `/riot` routes.
- `/ws` remains an internal sidecar-to-Tauri bridge endpoint, not a frontend dependency.
- Existing config and log behavior.

Phase 1 does change how the app is launched and packaged:

- Tauri opens a native desktop window instead of launching the system browser.
- Tauri starts and supervises the Go sidecar.
- Tauri Rust connects to the sidecar `/ws` endpoint and emits typed app events to Vue.
- Frontend assets are built by Vite and consumed by Tauri.
- The Go binary is built as a sidecar without embedding frontend assets.
- The legacy Go update check is removed.

## Files And Responsibilities

- Create `src-tauri/`: Tauri Rust application, config, permissions, icons, and sidecar launch code.
- Modify `frontend/package.json`: add Tauri scripts and keep the existing Vite scripts.
- Modify `frontend/vite.config.js`: make Vite output compatible with Tauri and keep current alias behavior.
- Modify `frontend/src/utils/request.js`: make backend base resolution work inside Tauri WebView and browser dev mode.
- Modify `frontend/src/websocket/index.js`: replace direct WebSocket usage with Tauri event listening and keep a browser-dev fallback.
- Modify `configs/config.go`: add an optional browser auto-open override or ensure packaged sidecar defaults do not open an external browser.
- Modify `cmd/shield/main.go`: remove the current update check from startup.
- Modify `scripts/build.sh`: split frontend embed build from sidecar build so Tauri can package the Go executable cleanly.
- Create `scripts/build-tauri-sidecar.ps1`: Windows-friendly sidecar build script for Tauri packaging.
- Create `docs/tauri-phase1.md`: developer notes for local dev, packaging, and release limitations.

## Key Decisions

- Use fixed backend port `9365` for phase 1 because the existing config already defaults to `:9365`.
- Keep HTTP for request/response APIs in phase 1.
- Replace frontend WebSocket transport with Tauri events. Rust owns the sidecar WebSocket connection and emits `shield-status`.
- Disable `web.auto_open` for the sidecar when launched by Tauri.
- Keep administrator elevation in the Go process for phase 1, then revisit whether Tauri should request elevation in a later phase.
- Remove current Go update check in phase 1. Tauri updater can be added later as a separate signed-release feature.

## Task 1: Baseline Verification

**Files:**
- Read: `frontend/package.json`
- Read: `scripts/build.sh`
- Read: `cmd/shield/main.go`
- Read: `configs/config.go`

- [ ] **Step 1: Confirm frontend dependencies install**

Run:

```powershell
pnpm --dir frontend install --frozen-lockfile
```

Expected: command exits with code `0`. If `pnpm` is unavailable, run:

```powershell
npm --prefix frontend ci
```

Expected: command exits with code `0`.

- [ ] **Step 2: Confirm frontend builds before Tauri changes**

Run:

```powershell
pnpm --dir frontend build
```

Expected: `frontend/dist` is created and Vite exits with code `0`.

- [ ] **Step 3: Confirm Go tests pass before Tauri changes**

Run:

```powershell
go test ./...
```

Expected: all existing packages pass. If a test already fails before changes, record the package and failure text in the task notes before continuing.

- [ ] **Step 4: Confirm current Go binary builds**

Run:

```powershell
go build -tags jsoniter -ldflags "-s -w" -o bin/lol-shield.exe ./cmd/shield/main.go
```

Expected: `bin/lol-shield.exe` exists.

## Task 2: Scaffold Tauri App

**Files:**
- Create: `src-tauri/Cargo.toml`
- Create: `src-tauri/tauri.conf.json`
- Create: `src-tauri/src/main.rs`
- Create: `src-tauri/src/lib.rs`
- Create: `src-tauri/capabilities/default.json`
- Modify: `frontend/package.json`

- [ ] **Step 1: Add Tauri frontend scripts**

Modify `frontend/package.json` scripts to include:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "tauri": "tauri",
    "tauri:dev": "tauri dev",
    "tauri:build": "tauri build"
  }
}
```

Expected: existing scripts remain available.

- [ ] **Step 2: Create minimal Tauri Rust crate**

Create `src-tauri/Cargo.toml`:

```toml
[package]
name = "lol-shield-tauri"
version = "1.0.0"
description = "LOL Shield desktop shell"
authors = ["LOL Shield"]
edition = "2021"

[lib]
name = "lol_shield_tauri_lib"
crate-type = ["staticlib", "cdylib", "rlib"]

[build-dependencies]
tauri-build = { version = "2", features = [] }

[dependencies]
tauri = { version = "2", features = [] }
tauri-plugin-shell = "2"
futures-util = "0.3"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
tokio = { version = "1", features = ["time"] }
tokio-tungstenite = "0.24"
```

- [ ] **Step 3: Create Tauri entrypoints**

Create `src-tauri/src/main.rs`:

```rust
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    lol_shield_tauri_lib::run();
}
```

Create `src-tauri/src/lib.rs` with sidecar startup and a Rust-owned WebSocket bridge:

```rust
use std::time::Duration;

use futures_util::StreamExt;
use tauri::Emitter;
use tauri_plugin_shell::ShellExt;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::tungstenite::Message;

const SIDECAR_WS_URL: &str = "ws://127.0.0.1:9365/ws";

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
                        Ok(Message::Close(_)) => break,
                        Err(_) => break,
                        _ => {}
                    }
                }
            }

            tokio::time::sleep(Duration::from_secs(3)).await;
        }
    });
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let sidecar = app
                .shell()
                .sidecar("lol-shield")
                .expect("failed to create lol-shield sidecar command");

            let (_rx, _child) = sidecar
                .spawn()
                .expect("failed to spawn lol-shield sidecar");

            start_status_bridge(app.handle().clone());

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

Expected: this starts the bundled sidecar and bridges sidecar status messages to the frontend as `shield-status` Tauri events.

- [ ] **Step 4: Create Tauri config**

Create `src-tauri/tauri.conf.json`:

```json
{
  "$schema": "https://schema.tauri.app/config/2",
  "productName": "LOL Shield",
  "version": "1.0.0",
  "identifier": "work.bigorange.lol-shield",
  "build": {
    "beforeDevCommand": "pnpm --dir ../frontend dev --host 127.0.0.1",
    "devUrl": "http://127.0.0.1:5173",
    "beforeBuildCommand": "pnpm --dir ../frontend build",
    "frontendDist": "../frontend/dist"
  },
  "app": {
    "windows": [
      {
        "title": "LOL Shield",
        "width": 1200,
        "height": 800,
        "minWidth": 960,
        "minHeight": 640,
        "resizable": true
      }
    ],
    "security": {
      "csp": null
    }
  },
  "bundle": {
    "active": true,
    "targets": ["nsis"],
    "externalBin": ["../bin/lol-shield"],
    "windows": {
      "nsis": {
        "installMode": "perMachine"
      }
    }
  }
}
```

Expected: Tauri builds Windows installer artifacts and includes the sidecar.

- [ ] **Step 5: Create default capability file**

Create `src-tauri/capabilities/default.json`:

```json
{
  "$schema": "../gen/schemas/desktop-schema.json",
  "identifier": "default",
  "description": "Default desktop permissions for LOL Shield",
  "windows": ["main"],
  "permissions": ["core:default", "shell:allow-open"]
}
```

- [ ] **Step 6: Run Tauri metadata check**

Run:

```powershell
pnpm --dir frontend tauri info
```

Expected: Tauri prints environment information without config parse errors.

## Task 3: Build Go Sidecar For Tauri

**Files:**
- Create: `scripts/build-tauri-sidecar.ps1`
- Modify: `cmd/shield/main.go`

- [ ] **Step 1: Add sidecar build script**

Create `scripts/build-tauri-sidecar.ps1`:

```powershell
$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$bin = Join-Path $root "bin"
New-Item -ItemType Directory -Force $bin | Out-Null

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -tags jsoniter -ldflags "-s -w" -o (Join-Path $bin "lol-shield-x86_64-pc-windows-msvc.exe") ./cmd/shield/main.go
go build -tags jsoniter -ldflags "-s -w" -o (Join-Path $bin "lol-shield.exe") ./cmd/shield/main.go
```

Expected: Tauri target triple sidecar and plain dev binary are both produced.

- [ ] **Step 2: Add environment override for browser auto-open**

Modify `cmd/shield/main.go` so the startup path can force browser auto-open off when launched as a sidecar:

```go
if os.Getenv("LOL_SHIELD_TAURI_SIDECAR") == "1" {
    viper.Set(configs.WebAutoOpen, false)
}
```

Place it after `configs.Init(*configPath)` and before `shield := client.NewShield()`.

Add import:

```go
"os"
```

Expected: browser opening can be disabled without changing `config.yaml`.

- [ ] **Step 3: Remove Go update check from startup**

Modify `cmd/shield/main.go`:

- Delete the `VersionInfo` type.
- Delete the `checkUpdate()` function.
- Delete this startup block:

```go
syslog.L.Infof("配置初始化完成,正在检查更新中...")
err := checkUpdate()
if err != nil {
    syslog.L.Errorf("检查更新失败:%v", err)
}
```

Replace it with:

```go
syslog.L.Infof("配置初始化完成")
```

Remove imports that become unused after deleting the update check:

```go
"encoding/json"
"io"
"net/http"
"strings"
"time"
```

Expected: app startup no longer performs remote version checks.

- [ ] **Step 4: Pass sidecar environment from Tauri**

Modify `src-tauri/src/lib.rs` sidecar setup:

```rust
let sidecar = app
    .shell()
    .sidecar("lol-shield")
    .expect("failed to create lol-shield sidecar command")
    .env("LOL_SHIELD_TAURI_SIDECAR", "1");
```

Expected: sidecar runs without opening the system browser.

- [ ] **Step 5: Verify sidecar build**

Run:

```powershell
.\scripts\build-tauri-sidecar.ps1
```

Expected:

- `bin/lol-shield.exe` exists.
- `bin/lol-shield-x86_64-pc-windows-msvc.exe` exists.

## Task 4: Make Frontend HTTP And Status Events Tauri-Safe

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/src/utils/request.js`
- Modify: `frontend/src/websocket/index.js`

- [ ] **Step 1: Add Tauri frontend API package**

Modify `frontend/package.json` dependencies to include:

```json
{
  "dependencies": {
    "@tauri-apps/api": "^2.0.0"
  }
}
```

Expected: Vue code can import `listen` from `@tauri-apps/api/event`.

- [ ] **Step 2: Centralize HTTP backend base resolution**

Modify `frontend/src/utils/request.js` to keep HTTP APIs on the Go sidecar in Tauri:

```js
const defaultBackendBase = 'http://127.0.0.1:9365'
const backendBase = (
  import.meta.env.VITE_BACK_URL ||
  (window.__TAURI_INTERNALS__ ? defaultBackendBase : '')
).replace(/\/$/, '')
```

Expected:

- Browser dev mode without `VITE_BACK_URL` still uses relative paths.
- Tauri mode without `VITE_BACK_URL` uses `http://127.0.0.1:9365`.

- [ ] **Step 3: Replace frontend WebSocket with Tauri event listening**

Replace `frontend/src/websocket/index.js` with a Tauri-first adapter:

```js
import { listen } from '@tauri-apps/api/event'

const defaultBackendBase = 'http://127.0.0.1:9365'
const backendBase = (
    import.meta.env.VITE_BACK_URL ||
    (window.__TAURI_INTERNALS__ ? defaultBackendBase : '')
).replace(/\/$/, '')

const wsUrl = (() => {
    const base = backendBase || window.location.origin
    const u = new URL(base, window.location.origin)
    const protocol = u.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${u.host}/ws`
})()

let unlistenStatus = null
let ws = null
let heartbeatTimer = null
let isReconnect = true
let storeRef = null

export const createWebSocket = async (store) => {
    storeRef = store

    if (window.__TAURI_INTERNALS__) {
        if (unlistenStatus) {
            unlistenStatus()
        }

        unlistenStatus = await listen('shield-status', (event) => {
            if (store) {
                store.commit('ws/setWsRes', event.payload ?? {})
            }
        })
        return
    }

    createBrowserWebSocket(store)
}

const createBrowserWebSocket = (store) => {
    if (!('WebSocket' in window)) {
        console.log('该浏览器不支持 WebSocket')
        return
    }

    ws = new WebSocket(wsUrl)

    ws.onopen = function () {
        console.log('已连接后端')
        startHeartbeat()
    }

    ws.onmessage = function (msg) {
        if (store) {
            store.commit('ws/setWsRes', JSON.parse(msg.data ?? '{}'))
        }
    }

    ws.onerror = function (e) {
        console.log('ws错误:', e)
    }

    ws.onclose = function () {
        console.log('已关闭后端连接')
        stopHeartbeat()

        if (store) {
            store.commit('ws/reset')
        }

        if (isReconnect) {
            setTimeout(function () {
                console.log('尝试重新连接后端')
                createBrowserWebSocket(storeRef)
            }, 3 * 1000)
        }
    }
}

function startHeartbeat(interval) {
    interval = interval || 30
    heartbeatTimer = setInterval(function () {
        sendPing({ op: 1 })
    }, interval * 1000)
}

const sendPing = (message) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message))
    }
}

function stopHeartbeat() {
    clearInterval(heartbeatTimer)
}
```

Expected:

- Tauri runtime receives status through `shield-status` events.
- Browser dev fallback can still connect directly to `/ws`.
- Existing Vuex mutation `ws/setWsRes` does not need to change.

- [ ] **Step 4: Verify frontend build**

Run:

```powershell
pnpm --dir frontend build
```

Expected: build exits with code `0`.

- [ ] **Step 5: Verify Tauri event behavior**

Run Tauri dev:

```powershell
pnpm --dir frontend tauri:dev
```

Expected:

- Frontend does not create a browser `WebSocket` in Tauri runtime.
- Rust connects to `ws://127.0.0.1:9365/ws`.
- Vuex receives status payloads from the `shield-status` Tauri event.

## Task 5: Local Tauri Dev Flow

**Files:**
- Modify: `src-tauri/tauri.conf.json`
- Modify: `src-tauri/src/lib.rs`

- [ ] **Step 1: Build sidecar before starting Tauri dev**

Run:

```powershell
.\scripts\build-tauri-sidecar.ps1
```

Expected: sidecar binaries exist under `bin`.

- [ ] **Step 2: Start Tauri dev**

Run:

```powershell
pnpm --dir frontend tauri:dev
```

Expected:

- A Tauri window opens.
- The Go sidecar starts.
- No external browser opens.
- The frontend loads from Vite dev server.

- [ ] **Step 3: Verify app status without LOL running**

Expected:

- App remains open.
- Backend continues polling for LCU.
- UI does not crash if LOL is not running.

- [ ] **Step 4: Verify app status with LOL running**

Expected:

- Sidecar discovers LCU token.
- Rust bridge receives sidecar `/ws` messages and emits `shield-status`.
- UI receives online status through Tauri events.
- User info and current rank APIs work.

## Task 6: Packaging Flow

**Files:**
- Modify: `src-tauri/tauri.conf.json`
- Modify: `scripts/build-tauri-sidecar.ps1`
- Modify: `frontend/package.json`

- [ ] **Step 1: Build frontend and sidecar**

Run:

```powershell
pnpm --dir frontend build
.\scripts\build-tauri-sidecar.ps1
```

Expected:

- `frontend/dist` exists.
- Tauri target triple sidecar exists.

- [ ] **Step 2: Build Tauri installer**

Run:

```powershell
pnpm --dir frontend tauri:build
```

Expected:

- NSIS installer is created under `src-tauri/target/release/bundle/nsis`.
- Packaged app opens a Tauri window.
- Packaged app starts the sidecar.

- [ ] **Step 3: Verify installed app behavior**

Expected:

- App starts from Start Menu or installer completion.
- No console window flashes in release mode.
- UAC behavior is understandable and no worse than the current Go exe.
- Closing the Tauri window terminates or releases the Go sidecar.

## Task 7: Sidecar Lifecycle Hardening

**Files:**
- Modify: `src-tauri/src/lib.rs`

- [ ] **Step 1: Store sidecar child process in managed state**

Replace the ignored `_child` with a managed state wrapper:

```rust
use std::sync::Mutex;
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;

struct SidecarState(Mutex<Option<CommandChild>>);
```

In `run()`:

```rust
.manage(SidecarState(Mutex::new(None)))
```

Inside setup:

```rust
let (_rx, child) = sidecar.spawn().expect("failed to spawn lol-shield sidecar");
let state = app.state::<SidecarState>();
*state.0.lock().expect("sidecar state poisoned") = Some(child);
start_status_bridge(app.handle().clone());
```

Expected: the child handle is retained for shutdown logic.

- [ ] **Step 2: Kill sidecar on app exit**

Add an `on_window_event` or `RunEvent::ExitRequested` handler that takes the child from state and calls `kill()`.

Expected: closing the desktop app does not leave an orphaned `lol-shield.exe`.

- [ ] **Step 3: Verify no orphan process remains**

Run Tauri dev, close the window, then run:

```powershell
Get-Process lol-shield -ErrorAction SilentlyContinue
```

Expected: no leftover sidecar process unless another instance was already running before the test.

## Task 8: Documentation

**Files:**
- Create: `docs/tauri-phase1.md`

- [ ] **Step 1: Document local development**

Create `docs/tauri-phase1.md` with:

```markdown
# Tauri Phase 1

Phase 1 packages the existing Vue frontend in Tauri and runs the current Go service as a sidecar.
Tauri owns live status delivery: Rust connects to the sidecar `/ws` endpoint and emits `shield-status` events to Vue.

## Local Development

1. Install frontend dependencies:
   `pnpm --dir frontend install --frozen-lockfile`
2. Build the Go sidecar:
   `.\scripts\build-tauri-sidecar.ps1`
3. Start Tauri:
   `pnpm --dir frontend tauri:dev`

## Packaging

1. Build frontend:
   `pnpm --dir frontend build`
2. Build sidecar:
   `.\scripts\build-tauri-sidecar.ps1`
3. Build installer:
   `pnpm --dir frontend tauri:build`

## Phase 1 Limitations

- LCU logic remains in Go.
- Frontend still talks to the sidecar through HTTP for request/response APIs.
- Frontend does not directly connect to WebSocket in Tauri runtime.
- The legacy Go update check is removed.
- Signed Tauri updater integration is deferred to a later phase.
```

- [ ] **Step 2: Verify docs are accurate**

Run each documented command once on Windows and update the doc if paths or script names differ.

## Task 9: Regression Testing

**Files:**
- Test only unless fixes are needed.

- [ ] **Step 1: Run Go tests**

Run:

```powershell
go test ./...
```

Expected: all Go tests pass.

- [ ] **Step 2: Run frontend build**

Run:

```powershell
pnpm --dir frontend build
```

Expected: Vite build succeeds.

- [ ] **Step 3: Run Tauri build**

Run:

```powershell
.\scripts\build-tauri-sidecar.ps1
pnpm --dir frontend tauri:build
```

Expected: installer build succeeds.

- [ ] **Step 4: Manual smoke test without LOL**

Expected:

- App opens.
- UI renders all routes: `/`, `/rank`, `/running`.
- App stays responsive while backend reports waiting/offline state.

- [ ] **Step 5: Manual smoke test with LOL**

Expected:

- LCU online state appears.
- Current user loads.
- Match history loads.
- Live game page does not crash outside active game.
- Auto accept/pick/ban settings save to config.

## Risks And Mitigations

- **Sidecar target triple naming:** Tauri requires sidecar binaries to match target naming conventions. The build script creates `lol-shield-x86_64-pc-windows-msvc.exe` for packaging and `lol-shield.exe` for direct testing.
- **Port conflict on `9365`:** Phase 1 keeps the existing default. If a second instance starts, the Go server should fail clearly; later phases can add single-instance behavior.
- **UAC behavior:** Go sidecar still self-elevates. Verify packaged app behavior carefully because elevation may spawn a separate elevated process.
- **Status bridge reconnects:** The Rust bridge must retry `ws://127.0.0.1:9365/ws` because the sidecar may take a moment to start and LOL may not be running.
- **Update behavior removed:** Phase 1 does not check for updates. Add signed Tauri updater later only if release distribution needs it.
- **Orphan sidecar:** Retain and kill the child process from Tauri on exit.

## Definition Of Done

- `pnpm --dir frontend build` succeeds.
- `go test ./...` succeeds or known pre-existing failures are documented.
- `.\scripts\build-tauri-sidecar.ps1` creates the sidecar binaries.
- `pnpm --dir frontend tauri:dev` opens the app and starts the sidecar.
- `pnpm --dir frontend tauri:build` creates a Windows installer.
- Go startup no longer performs update checks.
- Tauri runtime receives sidecar status through `shield-status` events.
- Closing the Tauri app does not leave an orphaned sidecar process.
- Manual smoke tests pass with and without LOL running.
