import test from 'node:test'
import assert from 'node:assert/strict'

import { buildItemDict } from './generate-item-dict.mjs'
import items from '../src/model/dicts/items.generated.js'
import dicts from '../src/model/dicts/index.js'

test('buildItemDict keeps only Icons2D icons and strips the prefix', () => {
    const dict = buildItemDict([
        { id: 3008, iconPath: '/lol-game-data/assets/ASSETS/Items/Icons2D/3008_GluttonousGreaves.png' },
        { id: 12345, iconPath: '/lol-game-data/assets/ASSETS/Spells/Icons2D/something.png' },
        { id: 99999, iconPath: 'https://external.example/icon.png' },
        { id: 88888 },
    ])
    assert.deepEqual(dict, { '3008': '3008_GluttonousGreaves.png' })
})

test('buildItemDict sorts ids numerically and stringifies keys', () => {
    const dict = buildItemDict([
        { id: 3008, iconPath: '/lol-game-data/assets/ASSETS/Items/Icons2D/3008_GluttonousGreaves.png' },
        { id: 99, iconPath: '/lol-game-data/assets/ASSETS/Items/Icons2D/99_test.png' },
        { id: 2510, iconPath: '/lol-game-data/assets/ASSETS/Items/Icons2D/2510_APFighterSheen.png' },
    ])
    assert.deepEqual(Object.keys(dict), ['99', '2510', '3008'])
})

test('generated dict renders current-season items reported as missing icons', () => {
    const expected = {
        2510: '2510_APFighterSheen.png',
        2512: '2512_ADCAllIn.png',
        2517: '2517_ADFighterOmnivamp.png',
        2520: '2520_ADAssassinGameEnder.png',
        3008: '3008_GluttonousGreaves.png',
        1086: '1086_Dorans_Bow.png',
        8010: '8010_BloodlettersCurse.png',
    }
    for (const [id, file] of Object.entries(expected)) {
        assert.equal(items[id], file, `item ${id} should map to ${file}`)
    }
})

test('generated dict keeps legacy items and Ornn upgrades renderable', () => {
    assert.equal(items['3031'], '3031_Marksman_T3_InfinityEdge.png')
    assert.equal(items['1054'], '1054_Dorans_Shield.png')
    assert.equal(items['223008'], items['3008'])
    assert.ok(Object.keys(items).length > 500, 'dict should cover the whole shop history')
})

test('dicts.getFeDict exposes the generated dict consumed by RankDetail', () => {
    const gameItem = dicts.getFeDict('gameItem')
    assert.equal(gameItem['2510'], '2510_APFighterSheen.png')
    assert.equal(gameItem['3031'], '3031_Marksman_T3_InfinityEdge.png')
})