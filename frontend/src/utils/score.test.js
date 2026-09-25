import test from 'node:test'
import assert from 'node:assert/strict'

import { computeMatchScores, extractScoreMetrics, scoreTier } from './score.js'

test('extractScoreMetrics clamps missing fields to zero and derives KDA/KP', () => {
    const metrics = extractScoreMetrics({ stats: { kills: 7, deaths: 2, assists: 9 } }, 10)
    assert.equal(metrics.kda, 8)
    assert.equal(metrics.damage, 0)
    assert.equal(metrics.tank, 0)
    assert.equal(metrics.heal, 0)
    assert.equal(metrics.gold, 0)
    assert.equal(metrics.kp, 1.6)
})

test('a lone participant receives the neutral score', () => {
    const scores = computeMatchScores([
        { participantId: 1, teamId: 100, stats: { kills: 5, deaths: 5, assists: 5 } },
    ])
    assert.equal(scores[1].score, 5)
    assert.equal(scores[1].mvp, true)
    assert.equal(scores[1].svp, false)
})

test('identical participants share the neutral score and MVP/SVP split by team result', () => {
    const stats = {
        kills: 5,
        deaths: 5,
        assists: 5,
        goldEarned: 9000,
        totalDamageDealtToChampions: 15000,
        totalDamageTaken: 15000,
        totalHeal: 0,
    }
    const scores = computeMatchScores(
        [
            { participantId: 1, teamId: 100, stats: { ...stats } },
            { participantId: 2, teamId: 200, stats: { ...stats } },
        ],
        [
            { id: 100, win: false },
            { id: 200, win: true },
        ],
    )
    assert.equal(scores[1].score, 5)
    assert.equal(scores[2].score, 5)
    assert.equal(scores[1].mvp, false)
    assert.equal(scores[1].svp, true)
    assert.equal(scores[2].mvp, true)
})

test('winner-side top score becomes MVP and loser-side top score becomes SVP', () => {
    const scores = computeMatchScores(
        [
            {
                participantId: 1,
                teamId: 100,
                stats: {
                    kills: 12,
                    deaths: 1,
                    assists: 14,
                    goldEarned: 15000,
                    totalDamageDealtToChampions: 32000,
                    totalDamageTaken: 25000,
                    totalHeal: 4000,
                },
            },
            {
                participantId: 2,
                teamId: 100,
                stats: {
                    kills: 3,
                    deaths: 6,
                    assists: 8,
                    goldEarned: 9000,
                    totalDamageDealtToChampions: 12000,
                    totalDamageTaken: 21000,
                    totalHeal: 500,
                },
            },
            {
                participantId: 3,
                teamId: 200,
                stats: {
                    kills: 9,
                    deaths: 4,
                    assists: 5,
                    goldEarned: 14000,
                    totalDamageDealtToChampions: 28000,
                    totalDamageTaken: 18000,
                    totalHeal: 200,
                },
            },
            {
                participantId: 4,
                teamId: 200,
                stats: {
                    kills: 2,
                    deaths: 8,
                    assists: 6,
                    goldEarned: 8000,
                    totalDamageDealtToChampions: 9000,
                    totalDamageTaken: 22000,
                    totalHeal: 100,
                },
            },
        ],
        [
            { id: 100, win: true },
            { id: 200, win: false },
        ],
    )
    assert.equal(scores[1].mvp, true)
    assert.equal(scores[1].score, 10)
    assert.equal(scores[3].svp, true)
    assert.equal(scores[3].score, 5.3)
    assert.ok(scores[1].score > scores[2].score)
    assert.ok(scores[3].score > scores[4].score)
})

test('missing stats rows still produce finite scores', () => {
    const scores = computeMatchScores([
        { participantId: 1, teamId: 100, stats: {} },
        { participantId: 2, teamId: 200 },
    ])
    for (const id of [1, 2]) {
        assert.ok(Number.isFinite(scores[id].score))
        assert.ok(scores[id].score >= 0 && scores[id].score <= 10)
    }
})

test('scoreTier maps scores to badge tiers', () => {
    assert.equal(scoreTier(9.5), 'gold')
    assert.equal(scoreTier(7.2), 'good')
    assert.equal(scoreTier(5), 'fair')
    assert.equal(scoreTier(2.1), 'low')
    assert.equal(scoreTier(NaN), 'none')
})