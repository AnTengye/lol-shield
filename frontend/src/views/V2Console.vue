<template>
    <a-space direction="vertical" style="width: 100%;">
        <a-card title="V2 System State" :loading="loadingState">
            <a-space>
                <a-button type="primary" @click="refreshState">刷新状态</a-button>
                <a-tag :color="stateColor">{{ state.status }}</a-tag>
                <span>Flow: {{ state.gameFlow }}</span>
                <span>Port: {{ state.port }}</span>
                <span>Watcher: {{ state.watcherStatus }}</span>
                <span>Reconnect: {{ state.reconnectCount }}</span>
            </a-space>
            <div v-if="state.lastError" style="margin-top: 12px; color: #cf1322;">
                {{ state.lastError }}
            </div>
            <div style="margin-top: 8px; color: #666;">
                LastFlowEventAt: {{ state.lastFlowEventAt || '-' }} | LastPollAt: {{ state.lastPollAt || '-' }}
            </div>
        </a-card>

        <a-card title="V2 Running Snapshot" :loading="loadingSnapshot">
            <a-space>
                <a-button @click="refreshSnapshot">刷新对局快照</a-button>
            </a-space>
            <pre class="json-view">{{ snapshotText }}</pre>
        </a-card>
    </a-space>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useStore } from 'vuex'
import { getRunningSnapshotV2 } from '@/v2/api/running'

const store = useStore()
const loadingState = ref(false)
const loadingSnapshot = ref(false)
const snapshotText = ref('{}')

const state = computed(() => ({
    status: store.getters['connectionV2/status'],
    gameFlow: store.getters['connectionV2/gameFlow'],
    port: store.getters['connectionV2/port'],
    lastError: store.getters['connectionV2/lastError'],
    watcherStatus: store.getters['connectionV2/watcherStatus'],
    reconnectCount: store.getters['connectionV2/reconnectCount'],
    lastFlowEventAt: store.getters['connectionV2/lastFlowEventAt'],
    lastPollAt: store.getters['connectionV2/lastPollAt'],
}))

const stateColor = computed(() => {
    if (state.value.status === 'online') return 'success'
    if (state.value.status === 'connecting') return 'processing'
    return 'error'
})

const refreshState = async () => {
    loadingState.value = true
    try {
        await store.dispatch('connectionV2/refresh')
    } finally {
        loadingState.value = false
    }
}

const refreshSnapshot = async () => {
    loadingSnapshot.value = true
    try {
        const res = await getRunningSnapshotV2()
        snapshotText.value = JSON.stringify(res.data || {}, null, 2)
    } catch (error) {
        snapshotText.value = JSON.stringify(error, null, 2)
    } finally {
        loadingSnapshot.value = false
    }
}

onMounted(async () => {
    await refreshState()
})
</script>

<style scoped>
.json-view {
    margin-top: 12px;
    padding: 12px;
    background: #f7f7f7;
    border-radius: 6px;
    max-height: 480px;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-all;
}
</style>
