# Mock LCU Usage

## What it gives you

`mock-lcu` lets `lol-shield` run against a fixture-backed local HTTP service instead of a live League Client. The current default scenario is a stable `InProgress` scene with match history, match detail, ranked data, current summoner data, and placeholder assets.

## Start the mock service

From the repo root:

```bash
go run ./cmd/mock-lcu -addr 127.0.0.1:19365
```

Optional:

```bash
go run ./cmd/mock-lcu -addr 127.0.0.1:20001 -scenario-dir internal/mocklcu/fixtures/default
```

## Enable mock mode in shield

Set these config values:

```yaml
mock_lcu:
  enabled: true
  base_url: http://127.0.0.1:19365
  scenario: default
```

Then start shield normally:

```bash
go run ./cmd/shield -c config.yaml
```

## Suggested developer loop

1. Start `mock-lcu`.
2. Start `shield` with `mock_lcu.enabled=true`.
3. Open the app and verify:
   - `/v1/history/:uid` returns paged history
   - `/v1/game/running` returns the in-progress scene
   - `/riot/*assets` resolves through the same backend path used in normal mode

## Current fixture notes

- Default fixtures are stored in `internal/mocklcu/fixtures/default`.
- Seed data comes from `20260507.log`.
- Some large LCU responses in that log were truncated, so a few fixtures use minimal realistic payloads to keep handler and page flows working.

## Verification commands

```bash
go test ./internal/mocklcu ./internal/core/lcuapi ./internal/client
go test ./cmd/mock-lcu
```
