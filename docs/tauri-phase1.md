# Tauri Phase 1

Phase 1 packages the existing Vue frontend in Tauri and runs the current Go service as a sidecar.
Tauri owns live status delivery: Rust connects to the sidecar `/ws` endpoint and emits `shield-status` events to Vue.

## Local Development

Prerequisites:

- Rust and Cargo from `rustup`.
- Visual Studio Build Tools with MSVC and Windows SDK.
- WebView2 Runtime.

1. Install frontend dependencies:
   `corepack pnpm --dir frontend install --frozen-lockfile`
2. Build the Go sidecar:
   `.\scripts\build-tauri-sidecar.ps1`
3. Start Tauri:
   `corepack pnpm --dir frontend tauri:dev`

## Packaging

1. Build frontend:
   `corepack pnpm --dir frontend build`
2. Build sidecar:
   `.\scripts\build-tauri-sidecar.ps1`
3. Build installer:
   `corepack pnpm --dir frontend tauri:build`

## Phase 1 Limitations

- LCU logic remains in Go.
- Frontend still talks to the sidecar through HTTP for request/response APIs.
- Frontend does not directly connect to WebSocket in Tauri runtime.
- The legacy Go update check is removed.
- Signed Tauri updater integration is deferred to a later phase.
