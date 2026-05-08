# Mock LCU Design

## Summary

Build a standalone local HTTP mock service that simulates the subset of Riot LCU APIs needed by this project, with fixtures checked into the repository and a default scenario focused on making the realtime page work end-to-end while also keeping the history page usable.

The mock service is meant to replace the need to launch a real League Client for most day-to-day feature development and debugging. It must be lightweight to start, deterministic, easy to extend, and close enough to the real API shape that the frontend proxy, backend handlers, and image loading paths can be exercised together.

## Goals

- Allow most frontend and backend development without a real LCU process.
- Exercise real HTTP paths rather than in-memory fake data only.
- Keep fixtures maintainable and reviewable in git.
- Make the realtime page open successfully in a default scenario.
- Keep the history page and game detail page functional in the same scenario.
- Support image and asset loading through the same `/riot` path style the app already uses.

## Non-Goals

- Full fidelity emulation of every Riot LCU endpoint.
- Runtime scene switching in the first version.
- Log replay as a live runtime dependency.
- A generic record-and-replay framework for arbitrary future APIs.

## Recommended Approach

### Option A: Standalone mock LCU HTTP service

Run a dedicated process such as `cmd/mock-lcu` that serves LCU-like endpoints from repository fixtures.

Pros:

- Exercises the real HTTP stack and URL paths.
- Keeps mock behavior separate from production handlers.
- Easy to run alongside frontend, backend, or nginx.
- Scales cleanly as more endpoints are added.

Cons:

- Slightly more setup than an in-process fake.
- Requires a small amount of config plumbing in the main app.

### Option B: Add mock routes into the main backend

Serve `/mock-lcu/*` from the existing shield process.

Pros:

- Less setup.
- Fewer binaries to manage.

Cons:

- Blurs production and test concerns.
- Does not mimic the real LCU boundary as well.
- Easier to accidentally couple business logic to mock-only behavior.

### Option C: Build a full scenario engine with state transitions

Create a richer simulator with runtime mutation and complex behavior.

Pros:

- Powerful long-term capability.

Cons:

- Too large for the current need.
- Higher implementation and maintenance cost.

### Recommendation

Choose Option A. It gives the best balance of realism, maintainability, and implementation scope.

## Architecture

The design has three primary parts.

### 1. Standalone mock server

Add `cmd/mock-lcu` as a small HTTP server listening on a configurable local address, for example `127.0.0.1:19365`.

Responsibilities:

- Load a selected fixture scenario at startup.
- Route incoming requests by LCU-compatible path and query.
- Return fixture JSON bodies and image bytes with stable response codes and content types.
- Provide deterministic responses with no dependency on a real Riot client.

### 2. Fixture-backed LCU service implementation

Refactor `internal/core/lcuapi` so that the app can use a generic HTTP-backed `Service` implementation that talks to either:

- the real Riot client base URL, or
- the mock LCU base URL.

This avoids hard-wiring the production package functions directly into every call path and lets the main app switch targets using configuration instead of invasive branching.

### 3. Explicit mock mode in shield startup

Add config-driven startup behavior in `cmd/shield` and related config loading:

- `mock_lcu.enabled`
- `mock_lcu.base_url`

When enabled:

- skip token and port discovery against the real LCU process
- initialize the LCU service against `mock_lcu.base_url`
- keep the rest of the app behavior unchanged as much as possible

## Default Scenario Scope

The first version uses a single scenario named `default`.

It must cover enough endpoints and data to support:

- realtime page opening in `InProgress`
- realtime page team display
- recent match history shown for players
- rank lookup for displayed players
- history page list loading
- game detail page loading
- profile icons, champion icons, spell icons, item icons, and loadscreen assets resolving through `/riot`

The default scenario does not need to support alternate gameflow phases such as champ select or postgame in v1.

## Fixture Layout

Store fixtures in a structured directory under the repo, not as raw log files.

Suggested layout:

```text
internal/mocklcu/
  fixtures/
    default/
      gameflow-phase.json
      gameflow-session.json
      current-summoner.json
      conversations.json
      conversation-messages/
        champ-select.json
      ranked-stats/
        <puuid>.json
      summoners/
        by-puuid/
          <puuid>.json
      match-history/
        products/
          <puuid>/
            beg-0-end-9.json
            beg-0-end-8.json
        games/
          <gameId>.json
      assets/
        DATA/
        ASSETS/
        v1/
```

Design rules:

- Organize by endpoint purpose, not by original log line.
- Keep filenames stable and readable.
- Allow multiple request windows for paginated history.
- Check image assets into fixture storage only for paths actually needed by current screens.

## Log Ingestion Strategy

The provided log file is the bootstrap source, not the runtime source of truth.

Workflow:

1. Extract sample JSON payloads from the log once.
2. Normalize them into endpoint-specific fixture files.
3. Commit the fixtures to the repository.
4. Run the mock server only from fixture files afterward.

This keeps the ongoing workflow simple and avoids reparsing log text every time the mock starts.

An optional helper tool may be added later to speed up extraction of new fixtures from additional logs, but the first version should not depend on that helper at runtime.

## Routing and Response Rules

The mock server must expose routes that match the LCU endpoints already used by the app, including:

- `/lol-gameflow/v1/gameflow-phase`
- `/lol-gameflow/v1/session`
- `/lol-chat/v1/conversations`
- `/lol-chat/v1/conversations/:id/messages`
- `/lol-summoner/v1/current-summoner`
- `/lol-match-history/v1/products/lol/:puuid/matches`
- `/lol-match-history/v1/games/:gameId`
- `/lol-ranked/v1/ranked-stats/:puuid`
- `/lol-game-data/assets/*`
- `/lol-hovercard/v1/friend-info/:puuid` only if current flows need it

Behavior rules:

- Match both path parameters and relevant query parameters.
- For paginated history, resolve fixtures by `begIndex` and `endIndex`.
- Return `404` for unknown paths.
- Return stable content types:
  - `application/json` for API payloads
  - image MIME type for assets based on extension
- For missing assets inside an otherwise supported fixture set, prefer returning a stable placeholder image over an empty body.

## Main App Integration

The current codebase uses `internal/core/lcuapi.Service` as the seam. Keep that seam and improve its implementations.

Target structure:

- keep the existing real-LCU implementation
- add an HTTP-base-url implementation usable for mock mode
- have `NewShield` and startup config choose the implementation explicitly

This avoids changing downstream business logic such as handlers, websocket helpers, and realtime assembly code.

## Testing Strategy

The work should be done with test-first discipline.

### Unit tests

Add tests for:

- fixture lookup by request path and query
- history pagination window selection
- asset path resolution and MIME selection
- config selection for mock mode

### Integration tests for mock server

Start the mock server in tests and verify:

- `gameflow-phase` returns `InProgress`
- `gameflow-session` returns player/team data
- history requests return expected payloads for known `begIndex/endIndex`
- game detail requests return expected data for a known `gameId`
- ranked stats requests return expected payloads for known `puuid`
- asset requests return non-empty content with correct MIME

### Application-level tests

Use mock mode to initialize the app and verify existing handlers can execute against the mock service for:

- `/v1/history/:uid`
- `/v1/game/:gameId`
- `/v1/game/running`
- `/v1/user`

## Error Handling

If the mock server cannot load the default scenario at startup, it must fail fast with a clear error showing which fixture path is missing or malformed.

For unknown requests:

- return `404`
- include a concise JSON error body for API endpoints

For malformed known fixture files:

- fail startup rather than serving partial or ambiguous data

This keeps failures deterministic and easier to debug.

## Operational Workflow

The expected local development workflow becomes:

1. Start `mock-lcu`.
2. Start the shield backend with `mock_lcu.enabled=true`.
3. Start frontend or nginx as usual.
4. Develop and debug without opening a real Riot client.

For future expansion:

- add more fixture files for new endpoints
- extend the router table
- add more scenarios later if needed

## Implementation Outline

Implementation should proceed in this order:

1. Introduce fixture loader and route matcher in `internal/mocklcu`.
2. Add failing tests for fixture-backed history and gameflow endpoints.
3. Build `cmd/mock-lcu` server to satisfy those tests.
4. Add mock-mode config and base-url-driven `lcuapi.Service`.
5. Add startup wiring in shield.
6. Add end-to-end handler tests against mock mode.

## Risks and Mitigations

### Risk: fixture sprawl

Mitigation:

- keep only endpoints actually used
- organize by endpoint and identifier
- avoid raw log dumps as permanent artifacts

### Risk: mismatch between real LCU and mock shape

Mitigation:

- derive initial fixtures directly from real logs
- keep JSON structures as close to actual responses as possible
- add regression tests using fields consumed by the app

### Risk: mock mode leaks into production flow

Mitigation:

- explicit config gate
- separate binary for mock server
- no production defaults changed when mock mode is off

## Success Criteria

The first version is successful when:

- a developer can run the app without a real LCU process
- the realtime page loads in a stable default `InProgress` scene
- the history page loads and drills into detail
- images resolve through the same `/riot` path style used in normal operation
- adding a new fixture-backed endpoint is straightforward and localized
