import test from 'node:test'
import assert from 'node:assert/strict'

import { isAramLikeQueue, resolveQueueName } from './queue.js'

test('resolveQueueName uses generated queue metadata before raw game mode fallback', () => {
    assert.equal(resolveQueueName({}, 2400, 'KIWI'), '海克斯大乱斗')
})

test('resolveQueueName still uses explicit runtime queue names before metadata', () => {
    assert.equal(resolveQueueName({}, 2400, 'ARAM: Mayhem'), 'ARAM: Mayhem')
})

test('resolveQueueName falls back to raw mode only when queue metadata is unavailable', () => {
    assert.equal(resolveQueueName({}, 999999, 'KIWI'), 'KIWI')
})

test('isAramLikeQueue treats ARAM Mayhem as an ARAM-style queue', () => {
    assert.equal(isAramLikeQueue(2400, '海克斯大乱斗'), true)
})
