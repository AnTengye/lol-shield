function loadPreferences() {
  try {
    return JSON.parse(localStorage.getItem('shield.preferences') || '{}')
  } catch {
    return {}
  }
}
export default {
  namespaced: true,
  state: () => ({
    preferences: {
      autoNavigate: false,
      showRank: true,
      showParty: true,
      ...loadPreferences(),
    },
    detailOpen: false,
    settingsDirty: false,
    cacheGeneration: 0,
    backendOnline: false,
    snapshots: {},
  }),
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
  },
}
