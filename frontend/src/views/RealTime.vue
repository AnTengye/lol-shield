<template>
    <div :style="{ backgroundImage: `url(${backgroundUrl})` }" class="backgroud-img">
        <a-spin :spinning="initLoading">
            <a-result v-if="!gameStarted" status="warning" title="等待游戏对局加载后，方能获取数据">
            </a-result>
            <div v-else>
                <a-row justify="space-around" v-for="(team, index) in teamInfo" :key="index">
                    <a-col :span="4" v-for="user in team">
                        <div class="container">
                            <a-image class="image" :width="180" :src="user.skinUrl" :preview="false"
                                style="opacity: 0.7;" />
                            <div class="overlay">
                                <a-row class="justift-center-container" style="padding-top: 10px;">
                                    <button class="player-name" :style="{ fontSize: '18px', color: user.teamColor }"
                                        @click.stop="openPlayerHistory(user)">
                                        <strong>{{ user.name.gameName }}</strong>
                                    </button>
                                    <button class="player-name tag-line" :style="{ fontSize: '14px', color: user.teamColor }"
                                        @click.stop="openPlayerHistory(user)">
                                        <strong>#{{ user.name.tagLine }}</strong>
                                    </button>
                                </a-row>
                                <a-row justify="space-around">
                                    <a-col :span="4" v-show="checkShow('段位')">
                                        <a-tooltip>
                                            <template #title>{{ user.rankText }}</template>
                                            <a-avatar size="large" shape="square"
                                                :src="getAssetsFile('rank/' + user.tier + '.png')"
                                                :alt="user.rankText"></a-avatar>
                                        </a-tooltip>
                                    </a-col>
                                    <a-col :span="14" class="justift-center-container" v-show="checkShow('段位')">
                                        <div style="font-size: 10px;"><strong>{{ user.rankText }}</strong></div>
                                    </a-col>
                                </a-row>
                                <a-row justify="space-around">
                                    <a-col :span="14" class="justift-center-container" v-show="checkShow('段位')">
                                        <div style="font-size: 10px;">
                                            <strong>近{{ user.totalGames }}场胜率：</strong>
                                            <strong :style="{ color: winRateColor(user.winRate) }">{{ user.winRate
                                                }}%</strong>
                                        </div>
                                    </a-col>
                                </a-row>
                                <a-row class="justift-center-container" v-show="checkShow('战绩')">
                                    <a-list :loading="initLoading" item-layout="horizontal" :data-source="user.history">
                                        <template #renderItem="{ item }">
                                            <div style="width: 100%;">
                                                <a-list-item :class="[backgroundColor(item.win)]"
                                                    style="padding-left: 12px;padding-right: 12px;">
                                                    <a-skeleton avatar :title="false" :loading="!!initLoading" active>
                                                        <a-list-item-meta :description="item.desc"
                                                            style="align-items: center;font-weight: bold">
                                                            <template #title>
                                                                {{ item.queue }}
                                                            </template>
                                                            <template #avatar>
                                                                <a-avatar :src="item.championIcon" />
                                                            </template>
                                                        </a-list-item-meta>
                                                    </a-skeleton>
                                                </a-list-item>
                                            </div>
                                        </template>
                                    </a-list>
                                </a-row>
                            </div>
                        </div>
                    </a-col>
                    <a-divider style="height: 1px; background-color: black" v-if="index < 1"><strong
                            style="color: white;">{{
                                displayQueueName }}</strong></a-divider>
                </a-row>
                <a-row>
                    <a-checkbox-group v-model:value="showState" :options="plainOptions"
                        style="padding-top: 10px;padding-left: 20px;">
                        <template #label="{ label }">
                            <span style="color: white">{{ label }}</span>
                        </template>
                    </a-checkbox-group>
                </a-row>
            </div>
            <a-drawer v-model:open="historyDrawerOpen" placement="right" :width="440" :title="historyDrawerTitle">
                <GameHistoryList :puuid="selectedPlayer?.puuid || ''" :page-size="20" :show-pagination="false"
                    :show-result-icon="false" :selectable="false" :auto-select-first="false" />
            </a-drawer>

        </a-spin>

    </div>

</template>
<script setup>
import { getAssetsFile } from '@/utils/getAssetsUrl.js'
import { getGameRunning, getMulGameRankHighest } from '@/api/bog'
import { buildRuntimeRiotAssetUrl } from '@/utils/backend'
import { onMounted, ref, watch, computed } from 'vue';
import { isAramLikeQueue, resolveQueueName } from '@/utils/queue'

import moment from 'moment';
import dicts from '@/model/dicts/index'
import { useStore } from 'vuex'
import GameHistoryList from './components/GameHistoryList.vue';
let { getters } = useStore()
const initLoading = ref(false);
const queueMap = dicts.getDict('queue');
const queueTypeMap = dicts.getDict('queueType');
const rankMap = dicts.getDict('rank')
// 客户端游戏状态
const status = computed(() => getters['ws/getGameStatus'])
const gameStarted = ref(false);
// 队伍数据
const teamColor = ['#FF6B6B', '#DC143C', '#00FA9A', '#FFA500'];
const teamInfo = ref([]);
const allPuuids = ref([])
// 加载皮肤数据
const skinMap = JSON.parse(localStorage.getItem('skins'))
// 地图数据
const queueId = ref(0);
const queueName = ref('');
const displayQueueName = computed(() => resolveQueueName(queueMap, queueId.value, queueName.value))
const backgroundUrl = computed(() => {
    return isAramLikeQueue(queueId.value, displayQueueName.value) ? '/aram.png' : '/classic.png';
})
// 多选
const plainOptions = ['战绩', '段位'];
const showState = ref(['战绩', '段位']);
const historyDrawerOpen = ref(false);
const selectedPlayer = ref(null);
const historyDrawerTitle = computed(() => {
    if (!selectedPlayer.value) {
        return '最近20场对局'
    }
    const name = selectedPlayer.value.name?.gameName || '未知玩家'
    const tagLine = selectedPlayer.value.name?.tagLine || ''
    return `${name}${tagLine ? '#' + tagLine : ''} 最近20场对局`
})
// 监听客户端游戏状态
watch(() => status.value, (newId, oldId) => {
    gameStarted.value = newId === 2
    if (gameStarted.value) {
        fetchRunningData()
    }
})
const fetchRunningData = () => {
    initLoading.value = true
    getGameRunning().then(res => {
        let blue = []
        let red = []
        let colorIndex = 1
        let preColorMap = {}
        let gameHistory = res.data.allGameHistory
        let nameMap = res.data.userNameMap
        let teamSkinMap = res.data.skinMap
        let allPuuid = []
        for (var i in res.data.preTeam) {
            let val = res.data.preTeam[i]
            if (val.length > 1) {
                val.forEach(element => {
                    preColorMap[element.puuid] = teamColor[colorIndex];
                });
                colorIndex++
            }
        }
        let f = function (element) {
            allPuuid.push(element.puuid)
            let history = []
            let wins = 0
            let total = 0
            let historyList = gameHistory[element.puuid]
            if (historyList !== undefined) {
                historyList.forEach(historyItem => {
                    history.push(formatHistoryItem(historyItem))
                    if (historyItem.win) {
                        wins++
                    }
                    total++
                });
                history = history.slice(0, 3)
            }
            if (total === 0) {
                total = 1
            }
            let nameInfo = {}
            if (nameMap[element.puuid] !== undefined) {
                nameInfo = nameMap[element.puuid]
            }
            let skinId = teamSkinMap[element.puuid]
            let skinUrl = buildRuntimeRiotAssetUrl('')
            if (skinId !== undefined && skinMap[skinId.skinId] !== undefined) {
                skinUrl = buildRuntimeRiotAssetUrl(skinMap[skinId.skinId].loadScreenPath)
            }
            return {
                summonerId: element.summonerId,
                name: nameInfo,
                skinUrl: skinUrl,
                puuid: element.puuid,
                teamColor: preColorMap[element.puuid],
                history: history,
                winRate: (wins / total * 100).toFixed(0),
                totalGames: total,
            }
        }
        if (res.data.selfTeamInfo.teamId === 100) {
            res.data.selfTeamInfo.userList.forEach(element => blue.push(f(element)));
            res.data.enemyTeamInfo.userList.forEach(element => red.push(f(element)));
        } else {
            res.data.selfTeamInfo.userList.forEach(element => red.push(f(element)));
            res.data.enemyTeamInfo.userList.forEach(element => blue.push(f(element)));
        }
        allPuuids.value = allPuuid
        teamInfo.value = [blue, red]
        gameStarted.value = true
        queueId.value = res.data.queueId
        queueName.value = res.data.queueName || ''
    }).finally(() => {
        initLoading.value = false
        if (allPuuids.value.length > 0) {
            getMulGameRankHighest(allPuuids.value).then(res => {
                let allPuuidRankedStatus = {}
                let temp = teamInfo.value
                res.data.forEach(element => {
                    allPuuidRankedStatus[element.puuid] = {
                        division: element.data.division,
                        tier: element.data.tier,
                        queueType: element.data.queueType
                    }
                })
                temp.forEach(element => {
                    element.forEach(item => {
                        if (allPuuidRankedStatus[item.puuid].tier !== '' && allPuuidRankedStatus[item.puuid].queueType !== 'RANKED_TFT') {
                            let rankInfo = allPuuidRankedStatus[item.puuid]
                            if (rankInfo.queueType === undefined || rankInfo.tier === '') {
                                item.rankText = '暂无排位数据'
                            } else {
                                item.rankText = queueTypeMap[rankInfo.queueType] + ' ' + rankMap[rankInfo.tier] + (rankInfo.division == 'NA' ? '' : rankInfo.division)
                            }
                            item.tier = rankInfo.tier.toLowerCase()
                        }
                    })
                })
                teamInfo.value = temp
            })
        }
    })
}
onMounted(() => {
    fetchRunningData()
});
const checkShow = (state) => {
    return showState.value.some(element => element === state);
}
const winRateColor = (winRate) => {
    if (winRate >= 70) {
        return 'red';
    } else if (winRate <= 40) {
        return 'green';
    } else {
        return 'white'; // 默认颜色
    }
}
const backgroundColor = (win) => {
    return win ? 'gradient-background-w' : 'gradient-background-l';
}
const formatHistoryItem = (historyItem) => {
    const championIcon = buildRuntimeRiotAssetUrl(`/v1/champion-icons/${historyItem.championId}.png`)
    const desc = moment(historyItem.createTime).format('MM-DD HH:mm') + ' ' + historyItem.kills + '-' + historyItem.deaths + '-' + historyItem.assists
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
const openPlayerHistory = (user) => {
    selectedPlayer.value = user
    historyDrawerOpen.value = true
}
</script>
<style scoped>
.backgroud-img {
    background-repeat: no-repeat;
    background-size: cover;
    background-position: center;
    height: 100%;
}

.real-time-container {
    position: relative;
    /* display: flex; */
    justify-content: center;
    align-items: center;
    /* width: 300px; */
}

.real-time-hero-image {
    display: block;
    width: 100%;
    height: auto;
    opacity: 0.7;
}

.container {
    position: relative;
    width: 180px;
}

.overlay {
    /* display: grid; */
    position: absolute;
    top: 0;
    left: 0;
    width: 180px;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    /* 半透明黑色背景 */
    /* display: flex; */
    /* justify-content: center; */
    /* align-items: center; */
    color: white;
    font-size: 20px;
    text-align: center;
    overflow: hidden;
    z-index: 2;
}

.justift-center-container {
    display: grid;
    place-items: center;
}

.player-name {
    cursor: pointer;
    display: block;
    width: 170px;
    margin: 0 auto 2px;
    padding: 3px 6px;
    border: 1px solid rgba(255, 255, 255, 0.45);
    border-radius: 4px;
    background: rgba(0, 0, 0, 0.42);
    line-height: 1.25;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    transition: background-color 0.15s ease, border-color 0.15s ease;
    position: relative;
    z-index: 3;
}

.player-name:hover {
    background: rgba(0, 0, 0, 0.72);
    border-color: rgba(255, 255, 255, 0.85);
}

.tag-line {
    margin-bottom: 6px;
}

.gradient-background-w {
    background-color: #8fd6a9;
    width: 200px;
    height: 60px;
    font-size: 20px;
    opacity: 0.78;
}


.gradient-background-l {
    background-color: #e39a9a;
    width: 200px;
    height: 60px;
    font-size: 20px;
    opacity: 0.78;
}
</style>
