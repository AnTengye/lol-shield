# Realtime Loading Cards Design

## Goal

Redesign the realtime match page so it feels closer to League of Legends' official loading screen while removing the current history overlay that blocks champion splash art and player information.

## Current Problem

`frontend/src/views/RealTime.vue` renders each player as a narrow champion card with rank, win rate, name, and recent match history all stacked inside the same overlay. The embedded history list consumes most of the card, making the screen visually noisy and causing severe occlusion.

## Chosen Direction

Use a loading-screen card layout:

- Keep two teams displayed as two horizontal rows of five vertical player cards.
- Keep champion splash art as the primary visual for each card.
- Move recent match history out of the card body.
- Show only fast-scan information on the card: rank, recent win rate, player name, tag line, and a history action.
- Use a gold-on-dark visual language with thin borders, smoky overlays, and restrained contrast to echo the League loading screen without introducing new assets.

## Player Card

Each card contains:

- Champion splash art using the existing `user.skinUrl`.
- Top stat strip with rank and recent win rate.
- Optional rank emblem when rank display is enabled.
- Bottom nameplate with game name and tag line.
- A compact "战绩" button that opens the existing history drawer.
- Team/pre-made color signal as a subtle accent border or glow.

The card must not render the inline three-match history list. The splash art should remain visible and recognizable.

## History Interaction

Clicking either the player nameplate or the "战绩" button opens the existing right drawer.

The drawer continues to use `GameHistoryList` with:

- `page-size` 20
- pagination hidden
- result icons hidden
- selection disabled
- no auto-select

This keeps the backend and history component behavior unchanged.

## Display Controls

The existing display controls for `战绩` and `段位` remain, but they should be restyled as a compact dark toolbar.

- Turning off `战绩` hides the history button only.
- Turning off `段位` hides rank text/emblem only.
- Card dimensions and spacing stay stable regardless of toggle state.

## Data Flow

No backend API changes are required.

`fetchRunningData()` continues to build `teamInfo` from `getGameRunning()` and enrich ranks through `getMulGameRankHighest()`. The existing `history` array may remain in the player object for future use, but the realtime card should not render it inline.

## Styling Constraints

- Use the existing Vue and Ant Design Vue stack.
- Avoid new dependencies.
- Keep styles scoped to `RealTime.vue`.
- Use fixed or responsive card dimensions so controls do not resize the layout.
- Keep text readable and clipped with ellipsis where needed.
- Preserve the classic/ARAM background image behavior.

## Verification

Implementation should be verified by:

- Running the frontend build.
- Opening the realtime page in a browser or app preview when possible.
- Checking that the splash art, rank, win rate, nameplate, and history button do not overlap.
- Checking that opening the history drawer still loads the selected player's recent matches.
- Checking both display toggles keep card layout stable.
