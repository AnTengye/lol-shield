# Desktop-Only Release Design

## Summary

Shift the project to a Tauri-only desktop release model and remove the browser-focused embedded frontend release path from the shipping workflow.

The new release pipeline should build only the assets needed for the desktop app, avoid redundant frontend builds, and stop opening a visible backend console window when the desktop app launches its sidecar backend.

## Goals

- Make Tauri the only supported release artifact.
- Remove release-time dependence on `internal/client/web/dist` as a shipped frontend path.
- Avoid building the same frontend twice during a tag release.
- Prevent the desktop launch flow from showing a backend console window to end users.
- Keep mock LCU and current backend behavior usable inside the desktop app.

## Non-Goals

- Reworking backend handler behavior unrelated to desktop startup.
- Removing every trace of embedded frontend code in one pass if some source-level fallback remains useful for local development.
- Changing the UI or functional scope of the desktop application itself.

## Recommended Approach

### Option A: Minimal release cleanup

Keep the current code structure but simplify release to rely on `tauri build` only, while also hiding the sidecar console window.

Pros:

- Smallest behavioral change.
- Reduces CI time immediately.

Cons:

- Leaves the repository with browser-era embedded frontend paths that are no longer part of the supported product story.

### Option B: Desktop-first cleanup

Treat Tauri as the only supported product, stop publishing or preparing the embedded browser frontend path in release automation, and make sidecar startup explicitly desktop-oriented.

Pros:

- Matches the intended product direction.
- Reduces release complexity and cognitive overhead.
- Keeps implementation scope reasonable.

Cons:

- Requires touching both CI and startup/runtime seams.

### Option C: Full browser-path deletion

Delete all embedded frontend serving code and all related assets immediately.

Pros:

- Most explicit end state.

Cons:

- Highest regression risk because some local tools, tests, or fallback startup paths may still assume embedded assets exist.
- Harder to merge safely with concurrent work.

### Decision

Choose Option C for this change set. The product direction is now explicitly desktop only, so the implementation should remove browser-era embedded frontend runtime paths and related release coupling instead of merely de-prioritizing them.

## Product Direction

The supported product becomes:

- Tauri desktop application
- bundled `lol-shield` sidecar backend

The unsupported product becomes:

- browser-facing packaged release flow based on embedded frontend files served from the Go binary

This means release automation, tagging expectations, and runtime assumptions should all be optimized for the Tauri desktop path first.

## Release Pipeline Design

The current release workflow performs:

1. a standalone frontend build
2. a copy into `internal/client/web/dist`
3. a Go sidecar build
4. a Tauri build, which triggers another frontend build through `beforeBuildCommand`

That repeats work unnecessarily.

The new release workflow should:

1. install frontend dependencies
2. build the Go sidecar binary
3. run `tauri build` as the single frontend-aware build step
4. publish only the Tauri NSIS artifact and checksums

Design rules:

- Do not sync `frontend/dist` into `internal/client/web/dist` during release.
- Do not run a separate explicit `pnpm --dir frontend build` before `tauri build`.
- Keep the Tauri config as the source of truth for how desktop assets are prepared.

## Embedded Frontend Strategy

`internal/client/web/dist` should no longer be part of the release contract or the supported runtime path.

For this pass:

- remove embedded frontend serving from the supported runtime path
- stop treating embedded frontend files as a release artifact
- do not copy production frontend output into that directory in CI
- remove placeholder or fallback files if they are no longer required after code cleanup

This makes the supported architecture explicit instead of carrying a second frontend delivery model forward.

## Sidecar Startup Design

The desktop app currently launches `lol-shield` as a sidecar, but users should not see a backend terminal window.

The launch behavior should be adjusted so that:

- desktop sidecar startup is hidden by default on Windows
- elevation flow preserves sidecar arguments and environment without surfacing a normal console window
- shutdown behavior remains unchanged from the user’s perspective

Implementation direction:

- keep using the Tauri sidecar model
- adjust Windows process creation and elevation behavior so the sidecar runs without a visible console
- avoid introducing a second launcher binary unless the first approach proves impossible

## Runtime Boundaries

The desktop app remains responsible for:

- starting the sidecar backend
- maintaining the websocket/status bridge
- stopping the sidecar on app close

The backend remains responsible for:

- serving local HTTP APIs
- serving `/riot` assets needed by the frontend
- using mock LCU mode when configured

No browser release path should be needed to support those responsibilities.

## Testing and Verification

Verification should cover:

- unit tests for any new sidecar/elevation argument handling
- release workflow sanity through local command parity where practical
- Tauri desktop startup check confirming no visible backend console window appears
- mock LCU desktop startup still functioning after the release-path cleanup

At minimum, the implementation should verify:

- Go tests for touched packages
- frontend build or `tauri build` entrypoint behavior
- one local desktop launch path

## Risks

### Risk: Hidden dependency on embedded frontend output

Some local commands or fallback paths may still assume `internal/client/web/dist` contains a real build.

Mitigation:

- identify and delete or rewrite those paths in the same change set
- verify touched tests and startup paths after removing CI sync

### Risk: Windows elevation still causes a visible window

Console visibility may come from ShellExecute or from the sidecar binary subsystem itself.

Mitigation:

- first try process-launch configuration changes
- if needed, adjust the built sidecar binary characteristics in a narrowly scoped follow-up inside the same implementation

### Risk: Over-deleting browser-era code

A broad cleanup could disturb unrelated local workflows.

Mitigation:

- remove release obligations first
- only delete runtime code that is clearly unused by the supported desktop path

## Scope

This design is intentionally limited to:

- release workflow simplification
- desktop-only artifact assumptions
- no-console sidecar startup
- deletion of browser-era embedded frontend runtime coupling

It does not include a broad architecture rewrite of the backend or frontend.
