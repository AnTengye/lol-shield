<template>
    <a-space direction="vertical" style="width: 100%;">
        <a-card title="Running V2" :loading="loading">
            <a-space>
                <a-button type="primary" @click="refresh">刷新</a-button>
                <a-switch v-model:checked="autoRefresh" />
                <span>自动刷新(3s)</span>
                <a-tag>{{ queueText }}</a-tag>
            </a-space>
            <div v-if="errorText" style="margin-top: 12px; color: #cf1322;">{{ errorText }}</div>
        </a-card>

        <a-row :gutter="16">
            <a-col :span="12">
                <a-card title="我方">
                    <a-list :data-source="snapshot.selfTeam || []" bordered>
                        <template #renderItem="{ item }">
                            <a-list-item>
                                <a-space direction="vertical" style="width: 100%;">
                                    <div><strong>{{ item.gameName }}</strong>#{{ item.tagLine }}</div>
                                    <div>段位: {{ rankText(item.highest) }}</div>
                                    <div>近5场:
                                        <span v-for="h in item.history" :key="h.gameId"
                                            :style="{ color: h.win ? '#237804' : '#a8071a', marginRight: '8px' }">
                                            {{ h.kills }}-{{ h.deaths }}-{{ h.assists }}
                                        </span>
                                    </div>
                                </a-space>
                            </a-list-item>
                        </template>
                    </a-list>
                </a-card>
            </a-col>
            <a-col :span="12">
                <a-card title="敌方">
                    <a-list :data-source="snapshot.enemyTeam || []" bordered>
                        <template #renderItem="{ item }">
                            <a-list-item>
                                <a-space direction="vertical" style="width: 100%;">
                                    <div><strong>{{ item.gameName }}</strong>#{{ item.tagLine }}</div>
                                    <div>段位: {{ rankText(item.highest) }}</div>
                                    <div>近5场:
                                        <span v-for="h in item.history" :key="h.gameId"
                                            :style="{ color: h.win ? '#237804' : '#a8071a', marginRight: '8px' }">
                                            {{ h.kills }}-{{ h.deaths }}-{{ h.assists }}
                                        </span>
                                    </div>
                                </a-space>
                            </a-list-item>
                        </template>
                    </a-list>
                </a-card>
            </a-col>
        </a-row>
    </a-space>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getRunningSnapshotV2 } from '@/v2/api/running'

const loading = ref(false)
const autoRefresh = ref(true)
const timer = ref(null)
const errorText = ref('')
const snapshot = ref({
    queueId: 0,
    queueName: '',
    selfTeam: [],
    enemyTeam: [],
})

const queueText = computed(() => {
    if (!snapshot.value.queueId) return '未在对局中'
    return `${snapshot.value.queueName || '未知队列'} (${snapshot.value.queueId})`
})

const rankText = (r) => {
    if (!r || !r.tier) return '暂无'
    return `${r.queueType || ''} ${r.tier} ${r.division || ''}`.trim()
}

const refresh = async () => {
    loading.value = true
    errorText.value = ''
    try {
        const res = await getRunningSnapshotV2()
        snapshot.value = res.data || {}
    } catch (error) {
        errorText.value = error?.message || error?.field || '获取失败'
    } finally {
        loading.value = false
    }
}

const stopTimer = () => {
    if (timer.value) {
        clearInterval(timer.value)
        timer.value = null
    }
}

const startTimer = () => {
    stopTimer()
    timer.value = setInterval(() => {
        refresh()
    }, 3000)
}

watch(() => autoRefresh.value, (enabled) => {
    if (enabled) {
        startTimer()
    } else {
        stopTimer()
    }
})

onMounted(async () => {
    await refresh()
    startTimer()
})

onBeforeUnmount(() => {
    stopTimer()
})
</script>
