import test from 'node:test'
import assert from 'node:assert/strict'

import { buildPartyView, summarizeParties } from './party.js'

const player = (puuid) => ({ puuid })

test('keeps a real premade group distinct from several solo players', () => {
    const players = ['a1', 'b1', 'c1', 'c2', 'd1'].map(player)
    const view = buildPartyView({
        alpha: [player('a1')],
        beta: [player('b1')],
        gamma: [player('c1'), player('c2')],
        delta: [player('d1')],
    }, players)

    assert.equal(view.byPuuid.c1.partyKey, view.byPuuid.c2.partyKey)
    assert.equal(view.byPuuid.c1.partyLabel, '组队 A')
    assert.equal(view.byPuuid.c1.partySize, 2)

    for (const puuid of ['a1', 'b1', 'd1']) {
        assert.equal(view.byPuuid[puuid].partyKind, 'solo')
        assert.equal(view.byPuuid[puuid].partyLabel, '单排')
    }

    const summaries = summarizeParties(players.map(({ puuid }) => ({
        puuid,
        ...view.byPuuid[puuid],
    })))
    assert.deepEqual(
        summaries.map(({ partyLabel, size }) => [partyLabel, size]),
        [['组队 A', 2], ['单排', 3]],
    )
})

test('does not infer solo status when the backend has no party evidence', () => {
    const view = buildPartyView({}, [player('a1'), player('b1')])

    assert.equal(view.hasEvidence, false)
    assert.equal(view.byPuuid.a1.partyKind, 'unknown')
    assert.equal(view.byPuuid.a1.partyLabel, '组队信息待确认')
    assert.deepEqual(view.summaries.map(({ partyLabel, size }) => [partyLabel, size]), [
        ['组队信息待确认', 2],
    ])
})
