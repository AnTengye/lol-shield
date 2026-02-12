# lol-shield v2 Architecture (LCU-First)

## 1. Scope

This v2 design targets a local-only, single-user desktop helper tightly coupled to Riot LCU APIs.

Non-goals:
- Multi-tenant auth
- Cloud deployment
- Backward compatibility with legacy internal module layout

## 2. Core Principles

1. LCU-first domain model: internal workflows are driven by LCU events and endpoint contracts.
2. Adapter isolation: LCU payload shape changes should be absorbed only in `internal/v2/lcu`.
3. Stable app API: frontend reads normalized view models, not raw LCU payloads.
4. Deterministic state machine: game phase transitions are centralized in one service.

## 3. Package Layout

```text
internal/v2/
  api/           # HTTP router and handlers for frontend
  app/           # state machine and orchestration service
  domain/        # pure models and transition logic
  lcu/           # LCU adapter and contracts
```

## 4. Runtime Pipeline

1. Engine polls local League client token/port.
2. LCU adapter initializes connection and obtains current game flow.
3. App state machine transitions and writes immutable snapshot into Store.
4. HTTP API returns current normalized state and config.

Future step: replace flow polling with websocket event ingestion while preserving same state machine.

## 5. API Contract (v2)

- `GET /api/v2/system/state`
  - Returns engine connection and game flow status.
- `GET /api/v2/config`
  - Returns local config slice used by automation.
- `PATCH /api/v2/config`
  - Updates local automation config and persists to `config.yaml`.
- `GET /api/v2/running/snapshot`
  - Returns normalized in-progress game snapshot with both teams, recent matches, ranked brief and skin binding.

## 6. Reliability Strategy

1. Single in-memory store guarded by RWMutex.
2. Clear status enum:
   - `offline`: League client not discovered
   - `connecting`: token found, validating LCU
   - `online`: LCU reachable and flow available
3. Poll loop with bounded interval and timeout.
4. Non-fatal errors are tracked in state snapshot for diagnostics.

## 7. LCU Maintainability Strategy

1. Keep endpoint constants and DTO mapping inside adapter package.
2. Add fixture replay tests in next tranche:
   - `fixtures/lcu/gameflow/*.json`
   - parser tests for each endpoint/event.
3. Any LCU format change should require adapter-only update.

## 8. Next Implementation Tranches

1. Add websocket subscriber and event dispatcher.
2. Add running snapshot aggregator (teams, ranks, skins).
3. Add frontend v2 pages based on stable API.
4. Add fixture-based contract tests and stress tests.
