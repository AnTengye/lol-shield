import test from 'node:test'
import assert from 'node:assert/strict'

import {
    describeUpdateError,
    formatBytes,
    isDesktopShell,
    isWindowsAgent,
    progressPercent,
} from './update.js'

test('isDesktopShell only accepts the tauri runtime', () => {
    assert.equal(isDesktopShell({ __TAURI_INTERNALS__: {} }), true)
    assert.equal(isDesktopShell({}), false)
    assert.equal(isDesktopShell(undefined), false)
    assert.equal(isDesktopShell(null), false)
})

test('isWindowsAgent matches windows user agents', () => {
    assert.equal(isWindowsAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64)'), true)
    assert.equal(isWindowsAgent('Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'), false)
    assert.equal(isWindowsAgent(''), false)
    assert.equal(isWindowsAgent(undefined), false)
})

test('formatBytes keeps download sizes readable', () => {
    assert.equal(formatBytes(0), '')
    assert.equal(formatBytes(-1), '')
    assert.equal(formatBytes(Number.NaN), '')
    assert.equal(formatBytes(512), '512 B')
    assert.equal(formatBytes(2048), '2.0 KB')
    assert.equal(formatBytes(12 * 1024 * 1024), '12.0 MB')
    assert.equal(formatBytes(150 * 1024 * 1024), '150 MB')
})

test('progressPercent clamps to 0-100 and tolerates unknown totals', () => {
    assert.equal(progressPercent(0, 0), 0)
    assert.equal(progressPercent(50, 0), 0)
    assert.equal(progressPercent(Number.NaN, 100), 0)
    assert.equal(progressPercent(1, 4), 25)
    assert.equal(progressPercent(4, 4), 100)
    assert.equal(progressPercent(5, 4), 100)
})

test('describeUpdateError explains missing platform packages', () => {
    assert.equal(describeUpdateError(''), '更新失败，请稍后重试。')
    assert.equal(describeUpdateError(new Error('')), '更新失败，请稍后重试。')
    assert.match(describeUpdateError('Targets not found: windows-x86_64'), /手动下载/)
    assert.equal(describeUpdateError(new Error('下载失败')), '下载失败')
    assert.equal(describeUpdateError('下载失败'), '下载失败')
})