<template>
    <div :style="{ backgroundImage: `url(${backgroundUrl})` }" class="backgroud-img">
        <a-spin :spinning="initLoading">
            <a-result v-if="!gameStarted" status="warning" title="等待游戏对局加载后，方能获取数据">
            </a-result>
            <div v-else class="match-board">
                <section class="team-row" v-for="(team, index) in teamInfo" :key="index">
                    <article class="player-card" v-for="user in team" :key="user.puuid"
                        :style="{ '--team-accent': user.teamColor || '#c8aa6e' }">
                        <a-image class="splash-image" :src="user.skinUrl" :preview="false" />
                        <div class="card-shade"></div>

                        <div class="stat-strip">
                            <div class="rank-pill" v-show="checkShow('段位')">
                                <a-avatar v-if="user.tier" class="rank-icon" size="small" shape="square"
                                    :src="getAssetsFile('rank/' + user.tier + '.png')" :alt="user.rankText" />
                                <span class="rank-text">{{ user.rankText || '暂无排位数据' }}</span>
                            </div>
                            <div class="winrate-pill">
                                <span>近{{ user.totalGames }}场</span>
                                <strong :style="{ color: winRateColor(user.winRate) }">{{ user.winRate }}%</strong>
                            </div>
                        </div>

                        <button class="player-nameplate" @click.stop="openPlayerHistory(user)">
                            <span class="game-name">{{ user.name.gameName || '未知玩家' }}</span>
                            <span class="tag-line" v-if="user.name.tagLine">#{{ user.name.tagLine }}</span>
                        </button>

                        <button v-show="checkShow('战绩')" class="history-action" @click.stop="openPlayerHistory(user)">
                            战绩
                        </button>
                    </article>

                    <div class="queue-divider" v-if="index < 1">
                        <span>{{ displayQueueName }}</span>
                    </div>
                </section>

                <div class="display-toolbar">
                    <a-checkbox-group v-model:value="showState" :options="plainOptions">
                        <template #label="{ label }">
                            <span class="display-option-label">{{ label }}</span>
                        </template>
                    </a-checkbox-group>
                </div>
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
        return '#f3d57a';
    } else if (winRate <= 40) {
        return '#7ed9a4';
    } else {
        return '#f4f0e6';
    }
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
    min-height: 100%;
    background-repeat: no-repeat;
    background-size: cover;
    background-position: center;
    height: 100%;
}

.match-board {
    min-height: 100%;
    padding: 24px 26px 18px;
    background:
        radial-gradient(circle at 50% 10%, rgba(200, 170, 110, 0.18), transparent 34%),
        linear-gradient(180deg, rgba(3, 10, 14, 0.34), rgba(2, 5, 8, 0.78));
}

.team-row {
    display: grid;
    grid-template-columns: repeat(5, minmax(130px, 1fr));
    gap: 14px;
    align-items: stretch;
    max-width: 1120px;
    margin: 0 auto 18px;
}

.player-card {
    --team-accent: #c8aa6e;
    position: relative;
    height: 260px;
    overflow: hidden;
    border: 1px solid rgba(200, 170, 110, 0.58);
    border-bottom-color: var(--team-accent);
    background: #071016;
    box-shadow: 0 16px 32px rgba(0, 0, 0, 0.34);
}

.player-card::before {
    content: '';
    position: absolute;
    inset: 0;
    border: 1px solid rgba(255, 255, 255, 0.08);
    pointer-events: none;
    z-index: 3;
}

.splash-image {
    width: 100%;
    height: 100%;
}

.splash-image :deep(.ant-image),
.splash-image :deep(img) {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: cover;
}

.card-shade {
    position: absolute;
    inset: 0;
    background:
        linear-gradient(180deg, rgba(0, 0, 0, 0.58), transparent 28%),
        linear-gradient(0deg, rgba(0, 0, 0, 0.86), transparent 46%);
    z-index: 1;
}

.stat-strip,
.player-nameplate,
.history-action {
    position: absolute;
    left: 10px;
    right: 10px;
    z-index: 1;
}

.stat-strip {
    top: 10px;
    display: grid;
    gap: 7px;
}

.rank-pill,
.winrate-pill,
.player-nameplate,
.history-action,
.display-toolbar {
    background: rgba(2, 8, 12, 0.72);
    color: #f4f0e6;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.65);
}

.rank-pill,
.winrate-pill {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 28px;
    padding: 4px 8px;
    border: 1px solid rgba(200, 170, 110, 0.42);
    font-size: 12px;
    line-height: 1.2;
}

.rank-icon {
    flex: 0 0 auto;
}

.rank-text {
    min-width: 0;
}

.rank-text,
.game-name,
.tag-line {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.player-nameplate {
    bottom: 44px;
    display: grid;
    place-items: center;
    min-height: 44px;
    padding: 6px 8px;
    border: 1px solid rgba(200, 170, 110, 0.52);
    border-radius: 0;
    line-height: 1.2;
    transition: border-color 0.16s ease, background-color 0.16s ease;
}

.history-action {
    bottom: 10px;
    height: 26px;
    padding: 0 10px;
    border: 1px solid rgba(200, 170, 110, 0.58);
    border-radius: 0;
    color: #f0d27a;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0;
}

.player-nameplate:hover,
.history-action:hover {
    border-color: rgba(240, 210, 122, 0.9);
    background: rgba(5, 18, 24, 0.92);
}

.game-name {
    font-size: 15px;
    font-weight: 700;
}

.tag-line {
    color: #c7d2d8;
    font-size: 11px;
}

.queue-divider {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 44px;
    color: #e6d19a;
    font-weight: 700;
}

.queue-divider::before,
.queue-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    max-width: 360px;
    background: linear-gradient(90deg, transparent, rgba(200, 170, 110, 0.66), transparent);
}

.queue-divider span {
    padding: 0 18px;
}

.display-toolbar {
    max-width: 1120px;
    margin: 18px auto 0;
    padding: 10px 12px;
    border: 1px solid rgba(200, 170, 110, 0.28);
    background: rgba(2, 8, 12, 0.64);
}

.display-option-label {
    color: #f4f0e6;
}

@media (max-width: 900px) {
    .match-board {
        padding: 18px 14px;
    }

    .team-row {
        grid-template-columns: repeat(5, minmax(110px, 1fr));
        gap: 10px;
        overflow-x: auto;
        padding-bottom: 6px;
    }

    .player-card {
        height: 230px;
    }
}
</style>
