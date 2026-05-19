# Realtime Loading Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the realtime match view as League-loading-style player cards with a non-occluding history action.

**Architecture:** Keep the change scoped to the frontend realtime view. Add one lightweight Node structure test that parses `RealTime.vue` text and guards the key behavior: no inline card history list, stable history drawer usage, and a dedicated history action. The implementation reuses existing data and the existing `GameHistoryList` drawer.

**Tech Stack:** Vue 3 SFC, Ant Design Vue, Vite, Node built-in test runner, existing `GameHistoryList`.

---

## File Structure

- Modify `frontend/package.json`: add a `test:structure` script using Node's built-in test runner.
- Create `frontend/src/views/RealTime.structure.test.mjs`: lightweight structural regression tests for the realtime view.
- Modify `frontend/src/views/RealTime.vue`: replace the crowded overlay layout with loading-screen cards, keep existing drawer logic, and restyle scoped CSS.

## Task 1: Add Realtime View Structure Test

**Files:**
- Modify: `frontend/package.json`
- Create: `frontend/src/views/RealTime.structure.test.mjs`

- [ ] **Step 1: Add a failing structure test**

Create `frontend/src/views/RealTime.structure.test.mjs`:

```js
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, 'RealTime.vue'), 'utf8');
const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] || '';

describe('RealTime loading-card layout', () => {
  it('does not render match history inline inside player cards', () => {
    assert.equal(
      template.includes(':data-source="user.history"'),
      false,
      'player cards should not embed the recent match list'
    );
  });

  it('keeps history in the existing drawer and exposes a dedicated history action', () => {
    assert.match(template, /<a-drawer[\s\S]*historyDrawerOpen/);
    assert.match(template, /<GameHistoryList[\s\S]*:page-size="20"/);
    assert.match(template, /class="[^"]*history-action[^"]*"/);
    assert.match(template, /@click\.stop="openPlayerHistory\(user\)"/);
  });

  it('uses stable loading-card structure for both teams', () => {
    assert.match(template, /class="[^"]*match-board[^"]*"/);
    assert.match(template, /class="[^"]*team-row[^"]*"/);
    assert.match(template, /class="[^"]*player-card[^"]*"/);
    assert.match(template, /class="[^"]*player-nameplate[^"]*"/);
  });
});
```

- [ ] **Step 2: Add the package script**

Update `frontend/package.json` scripts to include:

```json
"test:structure": "node --test src/views/RealTime.structure.test.mjs"
```

Keep the existing scripts unchanged.

- [ ] **Step 3: Run the test and verify RED**

Run:

```powershell
pnpm --dir frontend test:structure
```

Expected: FAIL because the current template still contains `:data-source="user.history"` and does not yet have `history-action`, `match-board`, `team-row`, `player-card`, or `player-nameplate`.

- [ ] **Step 4: Commit the failing test**

Run:

```powershell
git add frontend/package.json frontend/src/views/RealTime.structure.test.mjs
git commit -m "test: guard realtime loading card layout"
```

## Task 2: Replace Inline History Overlay With Loading Cards

**Files:**
- Modify: `frontend/src/views/RealTime.vue`

- [ ] **Step 1: Replace the realtime board template**

In `frontend/src/views/RealTime.vue`, replace the `v-else` content that currently starts with:

```vue
<div v-else>
    <a-row justify="space-around" v-for="(team, index) in teamInfo" :key="index">
```

with this card layout:

```vue
<div v-else class="match-board">
    <section class="team-row" v-for="(team, index) in teamInfo" :key="index">
        <article
            class="player-card"
            v-for="user in team"
            :key="user.puuid"
            :style="{ '--team-accent': user.teamColor || '#c8aa6e' }"
        >
            <a-image class="splash-image" :src="user.skinUrl" :preview="false" />
            <div class="card-shade"></div>

            <div class="stat-strip">
                <div class="rank-pill" v-show="checkShow('段位')">
                    <a-avatar
                        v-if="user.tier"
                        class="rank-icon"
                        size="small"
                        shape="square"
                        :src="getAssetsFile('rank/' + user.tier + '.png')"
                        :alt="user.rankText"
                    />
                    <span class="rank-text">{{ user.rankText || '暂无排位数据' }}</span>
                </div>
                <div class="winrate-pill">
                    <span>近{{ user.totalGames }}场</span>
                    <strong :style="{ color: winRateColor(user.winRate) }">{{ user.winRate }}%</strong>
                </div>
            </div>

            <button class="player-nameplate" @click.stop="openPlayerHistory(user)">
                <span class="game-name">{{ user.name.gameName || '未知玩家' }}</span>
                <span class="tag-line" v-if="user.name.tagLine">#{{ user.name.tagLine }}</span>
            </button>

            <button
                v-show="checkShow('战绩')"
                class="history-action"
                @click.stop="openPlayerHistory(user)"
            >
                战绩
            </button>
        </article>

        <div class="queue-divider" v-if="index < 1">
            <span>{{ displayQueueName }}</span>
        </div>
    </section>

    <div class="display-toolbar">
        <a-checkbox-group v-model:value="showState" :options="plainOptions">
            <template #label="{ label }">
                <span class="display-option-label">{{ label }}</span>
            </template>
        </a-checkbox-group>
    </div>
</div>
```

- [ ] **Step 2: Remove the old inline history markup**

Delete the old card body section that contains:

```vue
<a-list :loading="initLoading" item-layout="horizontal" :data-source="user.history">
```

Also remove old Ant Design grid wrappers inside player cards (`a-row`, `a-col`, `.overlay`, `.container`) from the realtime card template.

- [ ] **Step 3: Run the structure test and verify GREEN for template behavior**

Run:

```powershell
pnpm --dir frontend test:structure
```

Expected: PASS.

- [ ] **Step 4: Commit the template change**

Run:

```powershell
git add frontend/src/views/RealTime.vue
git commit -m "feat: switch realtime view to loading cards"
```

## Task 3: Restyle the Realtime Page

**Files:**
- Modify: `frontend/src/views/RealTime.vue`

- [ ] **Step 1: Replace outdated scoped styles**

In `frontend/src/views/RealTime.vue`, replace the scoped styles for `.backgroud-img`, `.container`, `.overlay`, `.player-name`, `.tag-line`, `.gradient-background-w`, and `.gradient-background-l` with:

```css
.backgroud-img {
    min-height: 100%;
    background-repeat: no-repeat;
    background-size: cover;
    background-position: center;
}

.match-board {
    min-height: 100%;
    padding: 24px 26px 18px;
    background:
        radial-gradient(circle at 50% 10%, rgba(200, 170, 110, 0.18), transparent 34%),
        linear-gradient(180deg, rgba(3, 10, 14, 0.34), rgba(2, 5, 8, 0.78));
}

.team-row {
    display: grid;
    grid-template-columns: repeat(5, minmax(130px, 1fr));
    gap: 14px;
    align-items: stretch;
    max-width: 1120px;
    margin: 0 auto;
}

.player-card {
    --team-accent: #c8aa6e;
    position: relative;
    height: 260px;
    overflow: hidden;
    border: 1px solid rgba(200, 170, 110, 0.58);
    border-bottom-color: var(--team-accent);
    background: #071016;
    box-shadow: 0 16px 32px rgba(0, 0, 0, 0.34);
}

.player-card::before {
    content: '';
    position: absolute;
    inset: 0;
    border: 1px solid rgba(255, 255, 255, 0.08);
    pointer-events: none;
    z-index: 3;
}

.splash-image {
    width: 100%;
    height: 100%;
}

.splash-image :deep(.ant-image),
.splash-image :deep(img) {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: cover;
}

.card-shade {
    position: absolute;
    inset: 0;
    background:
        linear-gradient(180deg, rgba(0, 0, 0, 0.58), transparent 28%),
        linear-gradient(0deg, rgba(0, 0, 0, 0.86), transparent 46%);
    z-index: 1;
}

.stat-strip {
    position: absolute;
    top: 10px;
    left: 10px;
    right: 10px;
    display: grid;
    gap: 7px;
    z-index: 2;
}

.rank-pill,
.winrate-pill {
    min-height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 4px 8px;
    border: 1px solid rgba(200, 170, 110, 0.42);
    background: rgba(2, 8, 12, 0.72);
    color: #f2e6c9;
    font-size: 12px;
    line-height: 1.2;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.65);
}

.rank-icon {
    flex: 0 0 auto;
}

.rank-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.player-nameplate {
    position: absolute;
    left: 10px;
    right: 10px;
    bottom: 42px;
    z-index: 2;
    min-height: 44px;
    display: grid;
    place-items: center;
    padding: 6px 8px;
    border: 1px solid rgba(200, 170, 110, 0.52);
    border-radius: 0;
    background: rgba(2, 8, 12, 0.76);
    color: #f4f0e6;
    line-height: 1.2;
    transition: border-color 0.16s ease, background-color 0.16s ease;
}

.player-nameplate:hover,
.history-action:hover {
    border-color: rgba(240, 210, 122, 0.9);
    background: rgba(5, 18, 24, 0.92);
}

.game-name,
.tag-line {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.game-name {
    font-size: 15px;
    font-weight: 700;
}

.tag-line {
    color: #c7d2d8;
    font-size: 11px;
}

.history-action {
    position: absolute;
    left: 10px;
    right: 10px;
    bottom: 10px;
    z-index: 2;
    height: 26px;
    padding: 0 10px;
    border: 1px solid rgba(200, 170, 110, 0.58);
    border-radius: 0;
    background: rgba(3, 12, 16, 0.74);
    color: #f0d27a;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0;
}

.queue-divider {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 44px;
    color: #e6d19a;
    font-weight: 700;
}

.queue-divider::before,
.queue-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    max-width: 360px;
    background: linear-gradient(90deg, transparent, rgba(200, 170, 110, 0.66), transparent);
}

.queue-divider span {
    padding: 0 18px;
}

.display-toolbar {
    max-width: 1120px;
    margin: 18px auto 0;
    padding: 10px 12px;
    border: 1px solid rgba(200, 170, 110, 0.28);
    background: rgba(2, 8, 12, 0.64);
}

.display-option-label {
    color: #f4f0e6;
}

@media (max-width: 900px) {
    .match-board {
        padding: 18px 14px;
    }

    .team-row {
        grid-template-columns: repeat(5, minmax(110px, 1fr));
        gap: 10px;
        overflow-x: auto;
        padding-bottom: 6px;
    }

    .player-card {
        height: 230px;
    }
}
```

- [ ] **Step 2: Run structure test**

Run:

```powershell
pnpm --dir frontend test:structure
```

Expected: PASS.

- [ ] **Step 3: Run frontend build**

Run:

```powershell
pnpm --dir frontend build
```

Expected: PASS and Vite emits a production bundle.

- [ ] **Step 4: Commit the styling change**

Run:

```powershell
git add frontend/src/views/RealTime.vue
git commit -m "style: polish realtime loading cards"
```

## Task 4: Browser Verification And Final Cleanup

**Files:**
- Read: `frontend/src/views/RealTime.vue`
- Read: `frontend/package.json`

- [ ] **Step 1: Start the frontend dev server**

Run:

```powershell
pnpm --dir frontend dev --host 127.0.0.1
```

Expected: Vite starts and prints a local URL, usually `http://127.0.0.1:5173/`.

- [ ] **Step 2: Open the app in a browser**

Open the Vite URL and navigate to the realtime route used by the app. If routing requires the app shell, use the visible navigation rather than guessing alternate URLs.

- [ ] **Step 3: Verify the visual acceptance points**

Check these conditions at desktop width:

- Two team rows render as five stable player cards each when realtime data is available.
- Splash art is visible behind the stat strip and nameplate.
- Rank, win rate, nameplate, and history button do not overlap.
- Clicking the nameplate opens the right history drawer.
- Clicking the "战绩" button opens the same drawer.
- Turning off `战绩` hides the button without resizing cards.
- Turning off `段位` hides rank text/emblem without resizing cards.

- [ ] **Step 4: Remove generated browser artifacts if any appear**

Run:

```powershell
git status --short
```

If browser tooling created `.playwright-mcp/`, remove it with:

```powershell
$target = Resolve-Path .playwright-mcp -ErrorAction Stop; $root = (Resolve-Path .).Path; if (-not $target.Path.StartsWith($root)) { throw "Refusing to remove outside workspace: $($target.Path)" }; Remove-Item -LiteralPath $target.Path -Recurse -Force
```

- [ ] **Step 5: Final verification**

Run:

```powershell
pnpm --dir frontend test:structure
pnpm --dir frontend build
git status --short
```

Expected: structure test passes, build passes, and only intentional source changes remain.
