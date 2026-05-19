---
name: update-lol-queue-modes
description: Use when League of Legends adds or renames a queue, game mode, gameMode code, queueId, ARAM variant, rotating mode, or Chinese display name in this project.
---

# Update LoL Queue Modes

## Overview

Keep mode display names driven by Riot queue metadata, with small Chinese overrides for product language. Never fix a leaked raw code like `KIWI` by only patching one UI call site.

## Workflow

1. Identify the evidence:
   - Capture the raw `queueId`, raw `gameMode`, and any LCU `queueName`.
   - Check Riot's queue source: `https://static.developer.riotgames.com/docs/lol/queues.json`.
   - Prefer `queueId` over `gameMode`; raw all-caps `gameMode` values are technical fallbacks.

2. Write or update a regression test first:
   - Add cases in `frontend/src/utils/queue.test.js`.
   - Test the real symptom, for example `resolveQueueName({}, 2400, 'KIWI')`.
   - Include ARAM-like classification when the mode should use ARAM styling/background.
   - Run `node --test src/utils/queue.test.js` from `frontend` and confirm the new test fails before editing production logic.

3. Update Chinese queue names:
   - Edit `NAME_OVERRIDES` in `frontend/scripts/generate-queue-metadata.mjs`.
   - Keep generated data out of hand edits unless network is unavailable and the change is urgent.
   - Use concise Chinese product names, e.g. `2400: '海克斯大乱斗'`.

4. Regenerate queue metadata:
   - Run `pnpm generate:queues` from `frontend`.
   - Confirm `frontend/src/model/dicts/queues.generated.js` contains the new or renamed `queueId`.

5. Update resolver logic only when the rule changes:
   - Main resolver: `frontend/src/utils/queue.js`.
   - Expected priority: explicit static Chinese map, human-readable LCU name, generated Riot metadata, raw fallback, then `队列 ${queueId}`.
   - Treat all-caps values such as `KIWI`, `CLASSIC`, or `STRAWBERRY` as raw mode codes unless no better metadata exists.

6. Verify broadly:
   - Run `node --test src/utils/queue.test.js src/utils/assets.test.js src/utils/backend.test.js` from `frontend`.
   - Run `pnpm build` from `frontend`.
   - Run `go test ./...` from the repo root if backend structs, mock LCU fixtures, or response fields were touched.

## Quick Reference

| Need | File or command |
| --- | --- |
| Chinese display names | `frontend/scripts/generate-queue-metadata.mjs` `NAME_OVERRIDES` |
| Generated Riot queue data | `frontend/src/model/dicts/queues.generated.js` |
| Display resolver | `frontend/src/utils/queue.js` |
| Regression tests | `frontend/src/utils/queue.test.js` |
| Refresh metadata | `pnpm generate:queues` |
| Frontend test command | `node --test src/utils/queue.test.js src/utils/assets.test.js src/utils/backend.test.js` |

## Common Mistakes

- Do not add a new mode only to `frontend/src/model/dicts/enum.js`; that static table is legacy input, not the durable source.
- Do not let an all-caps raw `gameMode` outrank generated metadata.
- Do not classify ARAM-like modes only by Chinese text; generated metadata has `aramLike`.
- Do not claim support is complete until tests and build have been freshly run.
