<template>
    <a-row>
        <a-col :span="6">
            <a-list class="demo-loadmore-list" :loading="initLoading" item-layout="horizontal" :data-source="list">
                <template #loadMore>
                    <div v-if="!initLoading && !loading"
                        :style="{ textAlign: 'center', marginTop: '12px', height: '32px', lineHeight: '32px' }">
                        <Pagination :total="totalPages" :current="currentPage" @page-change="handlePageChange" />
                    </div>
                </template>
                <template #renderItem="{ item }">
                    <a-list-item :class="[backgroundColor(item.win), { 'active': item.gameId === rankDetailGameId }]"
                        @click="goGame(item.gameId)">
                        <a-skeleton avatar :title="false" :loading="!!loading" active>
                            <a-list-item-meta :description="item.desc" style="align-items: center;">
                                <template #title>
                                    {{ item.queue }}
                                </template>
                                <template #avatar>
                                    <a-avatar :src="item.championIcon" />
                                </template>
                            </a-list-item-meta>
                            <div><icon-font v-if="item.win" :style="{ fontSize: '40px' }" type="icon-shengli1" />
                                <icon-font v-else :style="{ fontSize: '40px' }" type="icon-shibai_--copy" />
                            </div>
                        </a-skeleton>
                    </a-list-item>
                </template>
            </a-list>
        </a-col>
        <a-col :span="18">
            <RankDetail :gameId="rankDetailGameId" :puuid="curuuid" @checkout-puuid="checkoutUser" />
        </a-col>
    </a-row>
</template>

<script setup>
import { onMounted, ref, computed, watch } from 'vue';
import { getGameList } from '@/api/bog'
import dicts from '@/model/dicts/index'
import moment from 'moment';
import { createFromIconfontCN } from '@ant-design/icons-vue';
import { resolveQueueName } from '@/utils/queue'
import { useStore } from 'vuex'
import RankDetail from './RankDetail.vue';
import Pagination from './components/Pagination.vue';
const currentPage = ref(1);
const totalPages = ref(9);

const IconFont = createFromIconfontCN({
    scriptUrl: import.meta.env.VITE_ICON_URL,
});
const initLoading = ref(false);
const loading = ref(false);
const list = ref([]);
const queueMap = dicts.getDict('queue');
const rankDetailGameId = ref(0)
let { getters } = useStore()
const selfuuid = computed(() => getters['ws/getUuid'])
const curuuid = ref('')
curuuid.value = selfuuid.value
const fetchGameHistory = (puuid) => {
    if (puuid === '') return
    loading.value = true
    getGameList(puuid, currentPage.value - 1, totalPages.value).then(res => {
        const temp = []
        res.data.forEach(element => {
            loading.value = false;
            const championIcon = import.meta.env.VITE_BACK_URL + '/riot/v1/champion-icons/' + element.championId + '.png'
            const desc = moment(element.createTime).format('MM-DD HH:mm') + '  KDA:' + element.kills + '-' + element.deaths + '-' + element.assists
            temp.push({
                desc: desc,
                gameId: element.gameId,
                queue: resolveQueueName(queueMap, element.queueId, element.gameMode),
                championIcon: championIcon,
                win: element.win,
                assists: element.assists,
                kills: element.kills,
                deaths: element.deaths,
            })
        });
        if (temp.length !== 0) {
            rankDetailGameId.value = temp[0].gameId
        }
        list.value = temp
    }).finally(() => {
        loading.value = false
    })
}
onMounted(() => {
    fetchGameHistory(curuuid.value)
});

watch(() => selfuuid.value, (newId, oldId) => {
    curuuid.value = newId
    fetchGameHistory(newId)
})
const backgroundColor = (win) => {
    return win ? 'gradient-background-win' : 'gradient-background-lose';
}

const goGame = (gameId) => {
    rankDetailGameId.value = gameId
}

const handlePageChange = (page) => {
    currentPage.value = page;
    fetchGameHistory(curuuid.value)
};
const checkoutUser = (puuid) => {
    currentPage.value = 1
    curuuid.value = puuid
    fetchGameHistory(puuid)
};
</script>
<style type="text/css">
.icon {
    font-size: 24px;
    /* Adjust the size as needed */
}

.gradient-background-win {
    background: linear-gradient(to right, #e0f7e9 30%, #ffffff 70%);
    /* 调整颜色停靠点 */
}

.gradient-background-win.active {
    background: #e0f7e9;
}

.gradient-background-lose {
    background: linear-gradient(to right, #f7e0e0 30%, #ffffff 70%);
    /* 调整颜色停靠点 */
}

.gradient-background-lose.active {
    background: #f7e0e0
        /* 调整颜色停靠点 */
}
</style>
