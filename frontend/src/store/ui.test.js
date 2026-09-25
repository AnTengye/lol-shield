import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createStore } from 'vuex'

vi.mock('@tauri-apps/api/app', () => ({
  getVersion: vi.fn(async () => '2.0.0-rc.11'),
}))
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(async () => undefined),
}))
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn() }))
vi.mock('@tauri-apps/plugin-updater', () => ({ check: vi.fn() }))
vi.mock('@tauri-apps/plugin-process', () => ({
  relaunch: vi.fn(async () => undefined),
}))

import { getVersion } from '@tauri-apps/api/app'
import { invoke } from '@tauri-apps/api/core'
import { check } from '@tauri-apps/plugin-updater'
import { relaunch } from '@tauri-apps/plugin-process'
import ui from './ui'
import wsModule from './websocket'
import { listen } from '@tauri-apps/api/event'
import { createWebSocket, destroyWebSocket } from '../websocket'

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
  invoke.mockReset().mockResolvedValue(undefined)
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
    // 重启前先把本地服务带回来，避免 relaunch 失败后界面长期离线
    expect(invoke).toHaveBeenCalledWith('resume_sidecar')
    expect(
      invoke.mock.invocationCallOrder.at(-1),
    ).toBeLessThan(relaunch.mock.invocationCallOrder[0])
  })

  it('安装器启动失败时恢复本地服务并保留错误', async () => {
    useDesktop()
    const update = releaseUpdate()
    update.install.mockRejectedValue(new Error('installer launch failed'))
    check.mockResolvedValue(update)
    const store = newStore()
    await store.dispatch('ui/checkUpdate', { silent: true })
    expect(await store.dispatch('ui/installUpdate')).toBe(false)
    expect(invoke).toHaveBeenCalledWith('prepare_update')
    expect(invoke).toHaveBeenCalledWith('resume_sidecar')
    expect(store.state.ui.updateStatus).toBe('failed')
    expect(store.state.ui.updateError).toBe('installer launch failed')
  })

  it('下载失败不停止或重启本地服务', async () => {
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
    expect(invoke).not.toHaveBeenCalledWith('resume_sidecar')
  })

  it('后台退出失败时禁止启动安装器并恢复监督状态', async () => {
    useDesktop()
    const update = releaseUpdate()
    check.mockResolvedValue(update)
    invoke.mockImplementation(async (command) => {
      if (command === 'prepare_update') throw new Error('后台仍未退出')
    })
    const store = newStore()
    await store.dispatch('ui/checkUpdate')
    expect(await store.dispatch('ui/installUpdate')).toBe(false)
    expect(update.install).not.toHaveBeenCalled()
    expect(store.state.ui.updateError).toBe('后台仍未退出')
    expect(invoke).toHaveBeenCalledWith('resume_sidecar')
    expect(invoke.mock.calls.filter(([command]) => command === 'resume_sidecar')).toHaveLength(1)
  })
})

const sidecarSnapshot = (phase, revision = 1) => ({ phase, revision, message: '启动状态', logPath: 'C:\\logs' })
describe('本地服务生命周期', () => {
  afterEach(destroyWebSocket)

  async function eventStore() {
    useDesktop()
    const handlers = new Map()
    listen.mockImplementation(async (event, callback) => {
      handlers.set(event, callback)
      return () => handlers.delete(event)
    })
    const store = createStore({ modules: { ui, ws: { ...wsModule, state: () => ({ ...wsModule.state }) } } })
    await createWebSocket(store)
    return { store, emit: (event, payload) => handlers.get(event)?.({ payload }) }
  }

  it('停止后的晚到消息不能恢复在线状态，新启动也拒绝旧会话消息', async () => {
    const { store, emit } = await eventStore()
    const status = { Status: 1, GameStatus: 2, Uuid: 'player', Uid: 1 }
    emit('shield-sidecar', sidecarSnapshot('ready', 1))
    emit('shield-status', { revision: 1, status })
    expect(store.getters['ws/getStatus']).toBe(1)
    for (const phase of ['failed', 'stopping', 'updating']) {
      emit('shield-sidecar', sidecarSnapshot(phase, 2))
      emit('shield-status', { revision: 1, status })
      expect(store.getters['ws/getStatus']).toBe(0)
      expect(store.state.ui.backendOnline).toBe(false)
    }
    emit('shield-sidecar', sidecarSnapshot('ready', 3))
    emit('shield-status', { revision: 1, status })
    expect(store.getters['ws/getStatus']).toBe(0)
    emit('shield-status', { revision: 3, status })
    expect(store.getters['ws/getStatus']).toBe(1)
  })

  it('传输连接成功不代表受管后台就绪，旧监听器销毁后不再写状态', async () => {
    const { store, emit } = await eventStore()
    emit('shield-sidecar', sidecarSnapshot('failed'))
    emit('shield-transport', true)
    expect(store.state.ui.backendOnline).toBe(false)
    const lateListener = listen.mock.calls.find(([event]) => event === 'shield-sidecar')[1]
    destroyWebSocket()
    lateListener({ payload: sidecarSnapshot('ready', 2) })
    expect(store.state.ui.sidecar.phase).toBe('failed')
  })
  it('区分授权、失败和更新，旧快照不能覆盖新状态', () => {
    const store = newStore()
    store.commit('ui/sidecarState', sidecarSnapshot('authorizing', 1))
    expect(store.getters['ui/backendLabel']).toBe('等待管理员授权')
    store.commit('ui/sidecarState', sidecarSnapshot('failed', 3))
    store.commit('ui/sidecarState', sidecarSnapshot('ready', 2))
    expect(store.state.ui.backendOnline).toBe(false)
    expect(store.getters['ui/backendLabel']).toBe('本地服务启动失败')
    store.commit('ui/sidecarState', sidecarSnapshot('updating', 4))
    expect(store.getters['ui/backendLabel']).toBe('正在更新')
  })

  it('失败可主动重试，等待授权或更新时不重复启动', async () => {
    useDesktop()
    const store = newStore()
    for (const phase of ['starting', 'authorizing', 'ready', 'stopping', 'updating']) {
      store.commit('ui/sidecarState', sidecarSnapshot(phase))
      expect(await store.dispatch('ui/retrySidecar')).toBe(false)
    }
    expect(invoke).not.toHaveBeenCalled()
    store.commit('ui/sidecarState', sidecarSnapshot('failed'))
    invoke.mockImplementation(async (command) => command === 'sidecar_status' ? sidecarSnapshot('starting', 2) : undefined)
    expect(await store.dispatch('ui/retrySidecar')).toBe(true)
    expect(invoke).toHaveBeenCalledWith('resume_sidecar')
    expect(store.state.ui.sidecar.phase).toBe('starting')
  })

  it('读取管理状态失败会显示错误而非永久等待', async () => {
    useDesktop()
    invoke.mockRejectedValueOnce(new Error('IPC unavailable'))
    const store = newStore()
    await store.dispatch('ui/refreshSidecar')
    expect(store.state.ui.sidecarError).toContain('IPC unavailable')
  })
})