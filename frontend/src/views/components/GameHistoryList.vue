<template>
    <a-list class="game-history-list" :loading="loading" item-layout="horizontal" :data-source="list">
        <template #loadMore>
            <div v-if="showPagination && !loading"
                :style="{ textAlign: 'center', marginTop: '12px', height: '32px', lineHeight: '32px' }">
                <Pagination :total="totalPages" :current="currentPage" @page-change="handlePageChange" />
            </div>
        </template>
        <template #renderItem="{ item }">
            <a-list-item :class="[backgroundColor(item.win), { 'active': item.gameId === selectedGameId }]"
                @click="selectGame(item.gameId)">
                <a-skeleton avatar :title="false" :loading="!!loading" active>
                    <a-list-item-meta :description="item.desc" style="align-items: center;">
                        <template #title>
                            {{ item.queue }}
                        </template>
                        <template #avatar>
                            <a-avatar :src="item.championIcon" />
                        </template>
                    </a-list-item-meta>
                    <div v-if="showResultIcon">
                        <icon-font v-if="item.win" :style="{ fontSize: '40px' }" type="icon-shengli1" />
                        <icon-font v-else :style="{ fontSize: '40px' }" type="icon-shibai_--copy" />
                    </div>
                </a-skeleton>
            </a-list-item>
        </template>
    </a-list>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import { createFromIconfontCN } from '@ant-design/icons-vue';
import moment from 'moment';
import { getGameList } from '@/api/bog'
import dicts from '@/model/dicts/index'
import { buildRuntimeRiotAssetUrl } from '@/utils/backend'
import { resolveQueueName } from '@/utils/queue'
import Pagination from './Pagination.vue';

const props = defineProps({
    puuid: {
        type: String,
        default: ''
    },
    pageSize: {
        type: Number,
        default: 9
    },
    showPagination: {
        type: Boolean,
        default: true
    },
    showResultIcon: {
        type: Boolean,
        default: true
    },
    selectable: {
        type: Boolean,
        default: true
    },
    autoSelectFirst: {
        type: Boolean,
        default: true
    },
    selectedGameId: {
        type: Number,
        default: 0
    },
})

const emit = defineEmits(['game-change', 'page-loaded'])

const IconFont = createFromIconfontCN({
    scriptUrl: import.meta.env.VITE_ICON_URL,
});
const queueMap = dicts.getDict('queue');
const loading = ref(false);
const list = ref([]);
const currentPage = ref(1);
const totalItems = ref(0);
const hasNext = ref(false);
const totalPagesFromTotal = computed(() => Math.max(1, Math.ceil(totalItems.value / props.pageSize)));
const totalPages = computed(() => {
    if (hasNext.value) {
        return Math.max(totalPagesFromTotal.value, currentPage.value + 1)
    }
    return Math.max(totalPagesFromTotal.value, currentPage.value)
});

const formatHistoryItem = (historyItem) => {
    const championIcon = buildRuntimeRiotAssetUrl(`/v1/champion-icons/${historyItem.championId}.png`)
    const desc = moment(historyItem.createTime).format('MM-DD HH:mm') + '  KDA:' + historyItem.kills + '-' + historyItem.deaths + '-' + historyItem.assists
    return {
        desc: desc,
        gameId: historyItem.gameId,
        queue: resolveQueueName(queueMap, historyItem.queueId, historyItem.gameMode),
        championIcon: championIcon,
        win: historyItem.win,
        assists: historyItem.assists,
        kills: historyItem.kills,
        deaths: historyItem.deaths,
    }
}

const fetchGameHistory = () => {
    if (!props.puuid) return
    loading.value = true
    getGameList(props.puuid, currentPage.value - 1, props.pageSize).then(res => {
        const pageData = Array.isArray(res.data) ? {
            list: res.data,
            total: res.data.length,
            hasNext: res.data.length >= props.pageSize,
        } : res.data
        const games = pageData.list || []
        totalItems.value = pageData.total || 0
        hasNext.value = !!pageData.hasNext
        list.value = games.map(formatHistoryItem)
        emit('page-loaded', {
            ...pageData,
            currentPage: currentPage.value,
            totalPages: totalPages.value,
        })
        if (props.autoSelectFirst && list.value.length !== 0) {
            emit('game-change', list.value[0].gameId)
        }
    }).finally(() => {
        loading.value = false
    })
}

const selectGame = (gameId) => {
    if (!props.selectable) return
    emit('game-change', gameId)
}

const handlePageChange = (page) => {
    currentPage.value = page
    fetchGameHistory()
};

watch(() => props.puuid, () => {
    currentPage.value = 1
    totalItems.value = 0
    hasNext.value = false
    list.value = []
    fetchGameHistory()
}, { immediate: true })

defineExpose({ fetchGameHistory })

const backgroundColor = (win) => {
    return win ? 'gradient-background-win' : 'gradient-background-lose';
}
</script>

<style scoped>
.game-history-list :deep(.ant-list-item) {
    cursor: pointer;
    border-radius: 4px;
    margin-bottom: 6px;
}

.gradient-background-win {
    background: linear-gradient(to right, #8fd6a9 0%, #e6f7ed 72%);
}

.gradient-background-win.active {
    background: #74c995;
}

.gradient-background-lose {
    background: linear-gradient(to right, #e39a9a 0%, #f9e7e7 72%);
}

.gradient-background-lose.active {
    background: #d98282;
}
</style>
