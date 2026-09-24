import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createStore } from 'vuex'

vi.mock('@tauri-apps/api/app', () => ({
  getVersion: vi.fn(async () => '2.0.0-rc.11'),
}))
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(async () => undefined),
}))
vi.mock('@tauri-apps/plugin-updater', () => ({ check: vi.fn() }))
vi.mock('@tauri-apps/plugin-process', () => ({
  relaunch: vi.fn(async () => undefined),
}))

import { getVersion } from '@tauri-apps/api/app'
import { invoke } from '@tauri-apps/api/core'
import { check } from '@tauri-apps/plugin-updater'
import { relaunch } from '@tauri-apps/plugin-process'
import ui from './ui'

const windowsAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
const macAgent = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'

function releaseUpdate(version = '2.0.0-rc.12') {
  const download = vi.fn(async (onEvent) => {
    onEvent({ event: 'Started', data: { contentLength: 2048 } })
    onEvent({ event: 'Progress', data: { chunkLength: 1024 } })
    onEvent({ event: 'Progress', data: { chunkLength: 1024 } })
    onEvent({ event: 'Finished' })
  })
  const install = vi.fn(async () => undefined)
  return {
    version,
    currentVersion: '2.0.0-rc.11',
    date: '2026-09-24T00:00:00Z',
    body: '更新说明',
    download,
    install,
  }
}

function useDesktop(agent = windowsAgent) {
  window.__TAURI_INTERNALS__ = {}
  Object.defineProperty(window.navigator, 'userAgent', {
    value: agent,
    configurable: true,
  })
}

const newStore = () => createStore({ modules: { ui } })

beforeEach(() => {
  const cache = new Map()
  vi.stubGlobal('localStorage', {
    getItem: (key) => cache.get(key) ?? null,
    setItem: (key, value) => cache.set(String(key), String(value)),
    removeItem: (key) => cache.delete(key),
    clear: () => cache.clear(),
  })
  delete window.__TAURI_INTERNALS__
})

describe('版本更新', () => {
  it('浏览器模式静默检查不发请求也不提示', async () => {
    const store = newStore()
    expect(await store.dispatch('ui/checkUpdate', { silent: true })).toBe(null)
    expect(check).not.toHaveBeenCalled()
    expect(store.state.ui.updateError).toBe('')
    expect(store.getters['ui/updateVisible']).toBe(false)
  })

  it('浏览器模式显式检查给出说明', async () => {
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: false })
    expect(check).not.toHaveBeenCalled()
    expect(store.state.ui.updateError).toContain('桌面版')
  })

  it('桌面静默检查发现新版本并标记可见', async () => {
    useDesktop()
    check.mockResolvedValue(releaseUpdate())
    const store = newStore()
    const result = await store.dispatch('ui/checkUpdate', { silent: true })
    expect(result.version).toBe('2.0.0-rc.12')
    expect(getVersion).toHaveBeenCalled()
    expect(store.state.ui.appVersion).toBe('2.0.0-rc.11')
    expect(store.state.ui.updateStatus).toBe('idle')
    expect(store.state.ui.updateCheckedAt).toBeGreaterThan(0)
    expect(store.getters['ui/updateVisible']).toBe(true)
  })

  it('静默检查失败时保留已发现的更新', async () => {
    useDesktop()
    check.mockResolvedValueOnce(releaseUpdate())
    check.mockRejectedValueOnce(new Error('network down'))
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: true })
    await store.dispatch('ui/checkUpdate', { silent: true })
    expect(store.state.ui.update.version).toBe('2.0.0-rc.12')
    expect(store.state.ui.updateError).toBe('')
  })

  it('显式检查失败时说明原因', async () => {
    useDesktop()
    check.mockRejectedValue(new Error('Targets not found: windows-x86_64'))
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: false })
    expect(store.state.ui.updateError).toContain('手动下载')
  })

  it('稍后提示后不再打断，出现更新版本时重新提示', async () => {
    useDesktop()
    check.mockResolvedValueOnce(releaseUpdate())
    const first = newStore()
    await first.dispatch('ui/checkUpdate', { silent: true })
    first.dispatch('ui/dismissUpdate')
    expect(first.state.ui.dismissedUpdate).toBe('2.0.0-rc.12')
    expect(first.getters['ui/updateVisible']).toBe(false)
    expect(localStorage.getItem('shield.update.dismissed')).toBe('2.0.0-rc.12')

    check.mockResolvedValueOnce(releaseUpdate())
    const second = newStore()
    await second.dispatch('ui/checkUpdate', { silent: true })
    expect(second.getters['ui/updateVisible']).toBe(false)

    check.mockResolvedValueOnce(releaseUpdate('2.0.0-rc.13'))
    await second.dispatch('ui/checkUpdate', { silent: true })
    expect(second.getters['ui/updateVisible']).toBe(true)
  })

  it('一键更新先停本地服务再安装，Windows 不重复重启', async () => {
    useDesktop()
    const update = releaseUpdate()
    check.mockResolvedValue(update)
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: true })
    expect(await store.dispatch('ui/installUpdate')).toBe(true)

    expect(update.download).toHaveBeenCalled()
    expect(invoke).toHaveBeenCalledWith('prepare_update')
    expect(update.install).toHaveBeenCalled()
    expect(
      invoke.mock.invocationCallOrder[0],
    ).toBeLessThan(update.install.mock.invocationCallOrder[0])
    expect(relaunch).not.toHaveBeenCalled()
    expect(store.state.ui.updateStatus).toBe('idle')
    expect(store.state.ui.updateProgress).toBe(100)
  })

  it('非 Windows 安装结束后重启应用', async () => {
    useDesktop(macAgent)
    const update = releaseUpdate()
    check.mockResolvedValue(update)
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: true })
    await store.dispatch('ui/installUpdate')
    expect(relaunch).toHaveBeenCalled()
  })

  it('安装失败时恢复本地服务并保留错误', async () => {
    useDesktop()
    const update = releaseUpdate()
    update.download.mockRejectedValue(new Error('下载中断'))
    check.mockResolvedValue(update)
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: true })
    expect(await store.dispatch('ui/installUpdate')).toBe(false)
    expect(update.install).not.toHaveBeenCalled()
    expect(store.state.ui.updateStatus).toBe('failed')
    expect(store.state.ui.updateError).toBe('下载中断')
    expect(invoke).toHaveBeenCalledWith('resume_sidecar')
  })
})