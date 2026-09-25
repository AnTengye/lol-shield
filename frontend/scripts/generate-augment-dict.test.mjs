import test from 'node:test'
import assert from 'node:assert/strict'

import { buildAugmentDict } from './generate-augment-dict.mjs'
import augments from '../src/model/dicts/augments.generated.js'
import dicts from '../src/model/dicts/index.js'

test('buildAugmentDict strips the asset prefix and keeps display fields', () => {
    const dict = buildAugmentDict([
        {
            id: 1009,
            augmentNameId: 'ARAM_BuffBuddies',
            nameTRA: '霸符兄弟',
            augmentSmallIconPath: '/lol-game-data/assets/ASSETS/UX/Cherry/Augments/Icons/BuffBuddies_small.png',
            rarity: 'kSilver',
        },
        { id: 99999, augmentSmallIconPath: 'https://external.example/icon.png' },
        { id: 88888 },
    ])
    assert.deepEqual(dict, {
        '1009': {
            n: '霸符兄弟',
            i: '/ASSETS/UX/Cherry/Augments/Icons/BuffBuddies_small.png',
            r: 'kSilver',
        },
    })
})

test('buildAugmentDict falls back to augmentNameId when nameTRA is empty', () => {
    const dict = buildAugmentDict([
        {
            id: 42,
            augmentNameId: 'ARAM_Example',
            nameTRA: '',
            augmentSmallIconPath: '/lol-game-data/assets/ASSETS/UX/Kiwi/Augments/Icons/Example_small.png',
        },
    ])
    assert.equal(dict['42'].n, 'ARAM_Example')
    assert.equal(dict['42'].r, '')
})

test('buildAugmentDict sorts ids numerically and stringifies keys', () => {
    const dict = buildAugmentDict([
        { id: 1320, augmentSmallIconPath: '/lol-game-data/assets/ASSETS/UX/Kiwi/Augments/Icons/UpgradeCollector_small.png' },
        { id: 9, augmentSmallIconPath: '/lol-game-data/assets/ASSETS/UX/Cherry/Augments/Icons/Nine_small.png' },
        { id: 1195, augmentSmallIconPath: '/lol-game-data/assets/ASSETS/UX/Cherry/Augments/Icons/GiantSlayer_small.png' },
    ])
    assert.deepEqual(Object.keys(dict), ['9', '1195', '1320'])
})

test('generated dict covers augments used by ARAM Mayhem match records', () => {
    assert.equal(augments['1009'].n, '霸符兄弟')
    assert.equal(augments['1195'].n, '巨人杀手')
    assert.equal(augments['1320'].n, '升级：收集者')
    assert.ok(Object.keys(augments).length > 500, 'dict should cover the whole augment pool')
})

test('dicts.getFeDict exposes the generated augment dict consumed by RankDetail', () => {
    const augment = dicts.getFeDict('augment')
    assert.equal(augment['1009'].i, '/ASSETS/UX/Cherry/Augments/Icons/BuffBuddies_small.png')
    assert.equal(augment['1320'].i, '/ASSETS/UX/Kiwi/Augments/Icons/UpgradeCollector_small.png')
})