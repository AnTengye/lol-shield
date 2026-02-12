import { getSystemStateV2 } from '@/v2/api/system'

export const state = {
    status: 'offline',
    gameFlow: 'None',
    port: 0,
    updatedAt: '',
    lastError: '',
    watcherStatus: 'idle',
    reconnectCount: 0,
    lastFlowEventAt: '',
    lastPollAt: '',
}

export const actions = {
    async refresh({ commit }) {
        const res = await getSystemStateV2()
        commit('setState', res.data || {})
        return res
    }
}

export const mutations = {
    setState(state, payload) {
        state.status = payload.connectionStatus || 'offline'
        state.gameFlow = payload.gameFlow || 'None'
        state.port = payload.port || 0
        state.updatedAt = payload.updatedAt || ''
        state.lastError = payload.lastError || ''
        state.watcherStatus = payload.watcherStatus || 'idle'
        state.reconnectCount = Number(payload.reconnectCount || 0)
        state.lastFlowEventAt = payload.lastFlowEventAt || ''
        state.lastPollAt = payload.lastPollAt || ''
    },
    reset(state) {
        state.status = 'offline'
        state.gameFlow = 'None'
        state.port = 0
        state.updatedAt = ''
        state.lastError = ''
        state.watcherStatus = 'idle'
        state.reconnectCount = 0
        state.lastFlowEventAt = ''
        state.lastPollAt = ''
    }
}

export const getters = {
    status: (state) => state.status,
    gameFlow: (state) => state.gameFlow,
    port: (state) => state.port,
    lastError: (state) => state.lastError,
    watcherStatus: (state) => state.watcherStatus,
    reconnectCount: (state) => state.reconnectCount,
    lastFlowEventAt: (state) => state.lastFlowEventAt,
    lastPollAt: (state) => state.lastPollAt,
}

export default {
    namespaced: true,
    state,
    actions,
    mutations,
    getters,
}
