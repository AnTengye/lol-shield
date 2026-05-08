# Mock LCU Desktop Usage

## What it gives you

`mock-lcu` lets the Tauri desktop app run against a fixture-backed local HTTP service instead of a live League Client. The current default scenario is a stable `InProgress` scene with match history, match detail, ranked data, current summoner data, and placeholder assets.

## Start the mock service

From the repo root:

```bash
go run ./cmd/mock-lcu -addr 127.0.0.1:19365
```

Optional:

```bash
go run ./cmd/mock-lcu -addr 127.0.0.1:20001 -scenario-dir internal/mocklcu/fixtures/default
```

## Enable mock mode for the desktop sidecar

Set these config values:

```yaml
mock_lcu:
  enabled: true
  base_url: http://127.0.0.1:19365
  scenario: default
```

For desktop runs, the sidecar should receive the same values:

```bash
go run ./cmd/shield -c config.yaml --tauri-sidecar
```

## Suggested desktop developer loop

1. Start `mock-lcu`.
2. Build the desktop sidecar:
   `./scripts/build-tauri-sidecar.ps1`
3. Launch the desktop app:
   `corepack pnpm --dir frontend tauri:dev`
4. Verify inside the desktop-backed flow:
   - `/v1/history/:uid` returns paged history
   - `/v1/game/running` returns the in-progress scene
   - `/riot/*assets` resolves through the same backend path used in normal mode

## Backend-only verification path

If you only want to validate the local APIs without opening the desktop shell, run:

```bash
go run ./cmd/shield -c config.yaml --tauri-sidecar
```

Then verify the same `/v1/history/:uid`, `/v1/game/running`, and `/riot/*assets` endpoints against the local service.

## Current fixture notes

- Default fixtures are stored in `internal/mocklcu/fixtures/default`.
- Seed data comes from `20260507.log`.
- Some large LCU responses in that log were truncated, so a few fixtures use minimal realistic payloads to keep handler and page flows working.

## Verification commands

```bash
go test ./internal/mocklcu ./internal/core/lcuapi ./internal/client
go test ./cmd/mock-lcu
./scripts/build-tauri-sidecar.ps1
corepack pnpm --dir frontend tauri:build
```
