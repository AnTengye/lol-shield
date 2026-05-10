import test from 'node:test'
import assert from 'node:assert/strict'

import { buildGameItemIconPath } from './assets.js'

test('buildGameItemIconPath uses the canonical Riot asset casing', () => {
    assert.equal(
        buildGameItemIconPath('3031_Marksman_T3_InfinityEdge.png'),
        '/ASSETS/Items/Icons2D/3031_Marksman_T3_InfinityEdge.png',
    )
})
