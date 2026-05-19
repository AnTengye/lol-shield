import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const viewPath = path.join(__dirname, 'RealTime.vue')

async function loadTemplate() {
    const source = await readFile(viewPath, 'utf8')
    const templateStart = source.indexOf('<template>')
    const scriptStart = source.indexOf('<script setup>')
    assert.notEqual(templateStart, -1, 'RealTime.vue should contain a <template> block')
    assert.notEqual(scriptStart, -1, 'RealTime.vue should contain a <script setup> block')

    const template = source.slice(templateStart + '<template>'.length, scriptStart)
    const templateEnd = template.lastIndexOf('</template>')
    assert.notEqual(templateEnd, -1, 'RealTime.vue should close the outer <template> block')

    return template.slice(0, templateEnd)
}

test('RealTime template removes inline history list rendering', async () => {
    const template = await loadTemplate()
    assert.doesNotMatch(
        template,
        /:data-source="user\.history"/,
        'expected inline user.history list rendering to be removed from the loading card template'
    )
})

test('RealTime template keeps the history drawer backed by GameHistoryList', async () => {
    const template = await loadTemplate()
    assert.match(template, /<a-drawer\b[\s\S]*v-model:open="historyDrawerOpen"/)
    assert.match(template, /<GameHistoryList\b/)
    assert.match(template, /:page-size="20"/)
})

test('RealTime template exposes the loading card structure hooks', async () => {
    const template = await loadTemplate()
    for (const className of [
        'match-board',
        'team-row',
        'player-card',
        'player-nameplate',
        'history-action',
    ]) {
        assert.match(
            template,
            new RegExp(`class="[^"]*\\b${className}\\b`),
            `expected template to include .${className}`
        )
    }
})
