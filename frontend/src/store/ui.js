import {
  describeUpdateError,
  formatBytes,
  isDesktopShell,
  isWindowsAgent,
  progressPercent,
} from '@/utils/update'

const DISMISSED_KEY = 'shield.update.dismissed'

// 更新句柄带插件内部资源标识，进入响应式代理会破坏其私有字段，因此留在模块作用域。
let pendingUpdate = null
let checking = null

function loadPreferences() {
  try {
    return JSON.parse(localStorage.getItem('shield.preferences') || '{}')
  } catch {
    return {}
  }
}

function loadDismissed() {
  try {
    return localStorage.getItem(DISMISSED_KEY) || ''
  } catch {
    return ''
  }
}

function saveDismissed(version) {
  try {
    if (version) localStorage.setItem(DISMISSED_KEY, version)
    else localStorage.removeItem(DISMISSED_KEY)
  } catch {
    /* 存储不可用时只在本次会话忽略提示 */
  }
}

function desktopWindow() {
  return typeof window === 'undefined' ? null : window
}

export default {
  namespaced: true,
  state: () => ({
    preferences: {
      autoNavigate: false,
      showRank: true,
      showParty: true,
      autoCheckUpdate: true,
      ...loadPreferences(),
    },
    detailOpen: false,
    settingsDirty: false,
    cacheGeneration: 0,
    backendOnline: false,
    snapshots: {},
    appVersion: '',
    update: null,
    updateStatus: 'idle',
    updateProgress: 0,
    updateDetail: '',
    updateError: '',
    updateCheckedAt: 0,
    dismissedUpdate: loadDismissed(),
  }),
  getters: {
    updateBusy: (state) =>
      state.updateStatus === 'downloading' ||
      state.updateStatus === 'installing',
    updateVisible: (state) =>
      Boolean(state.update) &&
      state.dismissedUpdate !== state.update.version &&
      !state.detailOpen,
  },
  mutations: {
    preference(state, values) {
      Object.assign(state.preferences, values)
      try {
        localStorage.setItem(
          'shield.preferences',
          JSON.stringify(state.preferences),
        )
      } catch {
        /* 存储不可用时保留本次会话偏好 */
      }
    },
    detailOpen(state, value) {
      state.detailOpen = value
    },
    settingsDirty(state, value) {
      state.settingsDirty = value
    },
    backendOnline(state, value) {
      state.backendOnline = value
    },
    cacheCleared(state) {
      state.snapshots = {}
      state.cacheGeneration++
    },
    snapshot(state, { key, value }) {
      state.snapshots[key] = value
    },
    appVersion(state, value) {
      state.appVersion = value
    },
    updateState(state, values) {
      Object.assign(state, values)
    },
    updateChecked(state) {
      state.updateCheckedAt = Date.now()
    },
    dismissUpdate(state, version) {
      state.dismissedUpdate = version
      saveDismissed(version)
    },
  },
  actions: {
    async loadAppVersion({ state, commit }) {
      if (state.appVersion || !isDesktopShell(desktopWindow())) {
        return state.appVersion
      }
      try {
        const { getVersion } = await import('@tauri-apps/api/app')
        commit('appVersion', await getVersion())
      } catch {
        /* 取不到版本号时由界面回退到本地服务版本 */
      }
      return state.appVersion
    },
    async checkUpdate({ state, commit, dispatch }, { silent = true } = {}) {
      if (
        state.updateStatus === 'downloading' ||
        state.updateStatus === 'installing'
      ) {
        return null
      }
      if (!isDesktopShell(desktopWindow())) {
        if (!silent)
          commit('updateState', {
            updateError: '浏览器模式不检查更新，请在桌面版中操作。',
          })
        return null
      }
      if (checking) return checking
      commit('updateState', {
        updateStatus: 'checking',
        updateError: '',
        updateDetail: '',
      })
      checking = (async () => {
        try {
          await dispatch('loadAppVersion')
          const { check } = await import('@tauri-apps/plugin-updater')
          const result = await check()
          pendingUpdate = result || null
          const info = result
            ? {
                version: result.version,
                current: result.currentVersion,
                date: result.date || '',
                notes: result.body || '',
              }
            : null
          commit('updateState', { update: info })
          return info
        } catch (error) {
          // 静默检查失败时保留上一次结果，避免网络抖动抹掉已知的可用更新。
          if (!silent)
            commit('updateState', { updateError: describeUpdateError(error) })
          return null
        } finally {
          commit('updateChecked')
          commit('updateState', { updateStatus: 'idle' })
          checking = null
        }
      })()
      return checking
    },
    dismissUpdate({ state, commit }) {
      if (state.update) commit('dismissUpdate', state.update.version)
    },
    async installUpdate({ state, commit }) {
      const update = pendingUpdate
      if (
        !update ||
        state.updateStatus === 'downloading' ||
        state.updateStatus === 'installing'
      ) {
        return false
      }
      let received = 0
      let total = 0
      commit('updateState', {
        updateStatus: 'downloading',
        updateProgress: 0,
        updateError: '',
        updateDetail: '正在下载更新…',
      })
      try {
        await update.download((event) => {
          if (event.event === 'Started') {
            total = event.data?.contentLength || 0
            commit('updateState', {
              updateDetail: total
                ? `正在下载更新（${formatBytes(total)}）`
                : '正在下载更新…',
            })
          } else if (event.event === 'Progress') {
            received += event.data?.chunkLength || 0
            commit('updateState', {
              updateProgress: progressPercent(received, total),
            })
          } else if (event.event === 'Finished') {
            commit('updateState', {
              updateProgress: 100,
              updateDetail: '正在安装更新…',
            })
          }
        })
        const { invoke } = await import('@tauri-apps/api/core')
        commit('updateState', {
          updateStatus: 'installing',
          updateDetail: '正在安装更新，应用即将重启',
        })
        // 安装程序要替换安装目录里的 sidecar 可执行文件；sidecar 以管理员权限
        // 独立运行，prepare_update 会请求其自行退出并等待文件占用释放。
        await invoke('prepare_update').catch(() => {})
        try {
          await update.install()
        } catch (error) {
          // 安装器未接管退出（启动失败或用户取消 UAC）：恢复本地服务后再上报。
          await invoke('resume_sidecar').catch(() => {})
          throw error
        }
        // 走到这里说明进程未被安装器接管（非 Windows）：重启前先恢复本地服务。
        if (
          !isWindowsAgent(
            typeof navigator === 'undefined' ? '' : navigator.userAgent,
          )
        ) {
          await invoke('resume_sidecar').catch(() => {})
          const { relaunch } = await import('@tauri-apps/plugin-process')
          await relaunch()
        }
        commit('updateState', { updateStatus: 'idle', updateDetail: '' })
        return true
      } catch (error) {
        commit('updateState', {
          updateStatus: 'failed',
          updateDetail: '',
          updateError: describeUpdateError(error),
        })
        const { invoke } = await import('@tauri-apps/api/core')
        await invoke('resume_sidecar').catch(() => {})
        return false
      }
    },
  },
}
