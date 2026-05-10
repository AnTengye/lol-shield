import test from 'node:test'
import assert from 'node:assert/strict'

import {
    buildBackendUrl,
    buildRiotAssetUrl,
    getBackendBase,
} from './backend.js'

test('getBackendBase falls back to empty string for browser when env is unset', () => {
    assert.equal(getBackendBase({ envBase: '', isTauri: false }), '')
})

test('getBackendBase falls back to local backend for tauri when env is unset', () => {
    assert.equal(getBackendBase({ envBase: '', isTauri: true }), 'http://127.0.0.1:9365')
})

test('buildBackendUrl keeps relative api path when browser env base is unset', () => {
    assert.equal(buildBackendUrl('/v1/user', { envBase: '', isTauri: false }), '/v1/user')
})

test('buildRiotAssetUrl keeps relative riot path when browser env base is unset', () => {
    assert.equal(buildRiotAssetUrl('/v1/champion-icons/266.png', { envBase: '', isTauri: false }), '/riot/v1/champion-icons/266.png')
})

test('buildRiotAssetUrl prefixes tauri fallback backend when env is unset', () => {
    assert.equal(
        buildRiotAssetUrl('/ASSETS/Characters/Ahri/Skins/Base/AhriLoadscreen_0.jpg', { envBase: '', isTauri: true }),
        'http://127.0.0.1:9365/riot/ASSETS/Characters/Ahri/Skins/Base/AhriLoadscreen_0.jpg',
    )
})

test('buildBackendUrl trims trailing slash from explicit backend base', () => {
    assert.equal(
        buildBackendUrl('/v1/history/test', { envBase: 'http://localhost:9365/', isTauri: false }),
        'http://localhost:9365/v1/history/test',
    )
})
