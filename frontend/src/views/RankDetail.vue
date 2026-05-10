<template>
    <a-spin :spinning="initLoading">
        <a-row v-if="props.gameId !== 0" class="fade-total-background" style="height: 32px;" justify="space-around">
            <a-col class="justift-center-container" :span="4">
                <div>开始时间：{{ baseGameData.createTime }}</div>
            </a-col>
            <a-col class="justift-center-container" :span="4">
                <div>游戏模式：{{ baseGameData.queueId }}</div>
            </a-col>
            <a-col class="justift-center-container" :span="4">
                <div>游戏时长：{{ baseGameData.gameDuration }}min</div>
            </a-col>
        </a-row>
        <a-row v-for="team in teamInfo">
            <a-col :class="backgroundColor(team.teamId, baseGameData.winTeamId)" :span="24">
                <a-row style="height: 30px;" justify="space-around">
                    <a-col class="justift-center-container" :span="8">
                        <div class="defeat" v-if="baseGameData.winTeamId !== team.teamId">失败</div>
                        <div class="victory" v-else>胜利</div>
                    </a-col>
                    <a-col class="justift-center-container" :span="8">
                        <div style="font-size: 18px;"><icon-font :style="{ fontSize: '24px' }"
                                type="icon-caidanlan-xiaoshou-jingzhengduishou" />{{ team.kills }}
                        </div>
                    </a-col>
                    <a-col class="justift-center-container" :span="8">
                        <div style="font-size: 18px;"><icon-font :style="{ fontSize: '20px' }" type="icon-jinbi" />{{
                            formatNumber(team.goldEarned) }}
                        </div>
                    </a-col>
                </a-row>
                <a-row v-for="item in team.data" :key="item.puuid" justify="space-around">
                    <a-col :span="2" class="justift-center-container">
                        <a-badge class="badge-level" :count="item.champLevel" :offset="[-8, 35]" color="orange">
                            <a-avatar shape="square" :src="item.championIcon" size="large" alt="空"></a-avatar>
                        </a-badge>
                        <p style="margin: 32px 0 0 0;font-size: 16px;">{{ item.kills }}-{{ item.deaths }}-{{
                            item.assists }}</p>
                    </a-col>
                    <a-col :span="2" class="justift-center-container">
                        <a-tooltip>
                            <template #title>{{ item.gameName }}</template>
                            <div class="text-overflow-container"
                                :style="{ fontSize: '16px', color: isCurrentUser(item.puuid) ? 'red' : 'inherit', fontWeight: isCurrentUser(item.puuid) ? 'bold' : 'normal' }"
                                @click="userClick(item.puuid)">{{ item.gameName }}</div>
                        </a-tooltip>
                    </a-col>
                    <a-col :span="1" class="justift-center-container">
                        <a-tooltip>
                            <template #title>{{ item.rankText }}</template>
                            <a-avatar size="large" shape="square" :src="getAssetsFile('rank/' + item.tier + '.png')"
                                :alt="item.rankText"></a-avatar>
                        </a-tooltip>

                    </a-col>
                    <a-col :span="7" class="justift-center-container">
                        <a-row class="justift-left-element">
                            <a-col style="text-align: right;" :span="3">
                                <a-avatar class="small-items-container" :src="item.spell1Icon" alt="空"></a-avatar>
                                <a-avatar class="small-items-container" :src="item.spell2Icon" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item0" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item1" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item2" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item3" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item4" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item5" alt="空"></a-avatar>
                            </a-col>
                            <a-col :span="3">
                                <a-avatar class="square" shape="square" :src="item.item6" alt="空"></a-avatar>
                            </a-col>
                        </a-row>
                    </a-col>
                    <a-col :span="12">
                        <a-row justify="space-around" align="middle">
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.kda === allTotalMal.kda }">
                                    <div>KDA</div>
                                    <div>{{ item.kda.toFixed(1) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.goldEarned === allTotalMal.goldEarned }">
                                    <div>经济</div>
                                    <div>{{ formatNumber(item.goldEarned) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px; "
                                    :class="{ maxScore: item.totalDamageDealtToChampions === allTotalMal.totalDamageDealtToChampions }">
                                    <div>伤害</div>
                                    <div>{{ formatNumber(item.totalDamageDealtToChampions) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.totalDamageTaken === allTotalMal.totalDamageTaken }">
                                    <div>承伤</div>
                                    <div>{{ formatNumber(item.totalDamageTaken) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.damageDealtToObjectives === allTotalMal.damageDealtToObjectives }">
                                    <div>资源</div>
                                    <div>{{ formatNumber(item.damageDealtToObjectives) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.timeCCingOthers === allTotalMal.timeCCingOthers }">
                                    <div>控制</div>
                                    <div>{{ formatNumber(item.timeCCingOthers) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="2" align="middle">
                                <div style="margin-top: 10px; font-size: 14px;"
                                    :class="{ maxScore: item.visionScore === allTotalMal.visionScore }">
                                    <div>视野</div>
                                    <div>{{ formatNumber(item.visionScore) }}</div>
                                </div>
                            </a-col>
                            <a-col :span="10">
                                <a-flex gap="small">
                                    <a-tooltip v-if="allTotalMal.kda === item.kda">
                                        <template #title>MVP</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-crown" />
                                    </a-tooltip>
                                    <a-tooltip v-if="allTotalMal.goldEarned === item.goldEarned">
                                        <template #title>钱钱钱！！！</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-a-poker1" />
                                    </a-tooltip>
                                    <a-tooltip
                                        v-if="allTotalMal.totalDamageDealtToChampions === item.totalDamageDealtToChampions">
                                        <template #title>砍人最狠的</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-knief-2eps" />
                                    </a-tooltip>
                                    <a-tooltip v-if="allTotalMal.totalDamageTaken === item.totalDamageTaken">
                                        <template #title>被砍的最狠</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-armour" />
                                    </a-tooltip>
                                    <a-tooltip v-if="allTotalMal.assists === item.assists">
                                        <template #title>能蹭一个是一个</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-like" />
                                    </a-tooltip>
                                    <a-tooltip v-if="item.pentaKills === 1">
                                        <template #title>啊？五杀？</template>
                                        <icon-font :style="{ fontSize: '25px' }" type="icon-wusha" />
                                    </a-tooltip>
                                </a-flex>

                            </a-col>
                        </a-row>
                    </a-col>
                </a-row>
            </a-col>
        </a-row>
    </a-spin>

</template>
<script setup>
import { getAssetsFile } from '@/utils/getAssetsUrl.js'
import { onMounted, ref, watch, defineEmits } from 'vue';
import { getGameDetail, getMulGameRankHighest } from '@/api/bog'
import dicts from '@/model/dicts/index'
import moment from 'moment';
import { createFromIconfontCN } from '@ant-design/icons-vue';
import { buildRuntimeRiotAssetUrl } from '@/utils/backend'
import { buildRuntimeGameItemIconUrl } from '@/utils/assets.js'
import { resolveQueueName } from '@/utils/queue'
import { useStore } from 'vuex'
const IconFont = createFromIconfontCN({
    scriptUrl: import.meta.env.VITE_ICON_URL,
});
const props = defineProps(['gameId', 'puuid'])
const initLoading = ref(false);
const queueMap = dicts.getDict('queue');
const queueTypeMap = dicts.getDict('queueType');
const rankMap = dicts.getDict('rank')
const spellMap = dicts.getFeDict('spell');
const itemIconMap = dicts.getFeDict('gameItem');
const allPuuids = ref([])

// 定义基础数据
const baseGameData = ref({
    queueId: '自选匹配',
    createTime: '01-01 00:00',
    gameDuration: 60,
    winTeamId: 100,
})
const teamInfo = ref([])
const allTotalMal = ref({})
const partMap = ref([])
const fetchGameDetail = (gameId) => {
    if (gameId === 0) return
    initLoading.value = true
    getGameDetail(gameId).then(res => {
        let baseData = {}
        baseData.queueId = resolveQueueName(queueMap, res.data.queueId, res.data.gameMode)
        baseData.createTime = moment(res.data.gameCreation).format('MM-DD HH:mm')
        baseData.gameDuration = (res.data.gameDuration / 60).toFixed(1)
        res.data.teams.forEach(element => {
            if (element.win !== 'Fail') {
                baseData.winTeamId = element.teamId
            }
        });
        baseGameData.value = baseData
        res.data.participantIdentities.forEach(element => {
            if (element.player.summonerId === 0) {
                partMap[element.participantId] = {
                    puuid: element.player.puuid,
                    gameName: '人机',
                }
            } else {
                partMap[element.participantId] = {
                    puuid: element.player.puuid,
                    gameName: element.player.gameName + '#' + element.player.tagLine,
                }
            }

        })
        let totalPart = []
        let tempPart1 = []
        let tempPart2 = []
        let allPuuid = []
        let teamTotal1 = { teamId: 100, data: [], kills: 0, deaths: 0, assists: 0, goldEarned: 0 }
        let teamTotal2 = { teamId: 200, data: [], kills: 0, deaths: 0, assists: 0, goldEarned: 0 }
        res.data.participants.forEach(element => {
            let deaths = element.stats.deaths
            if (deaths === 0) {
                deaths = 1
            }
            let kda = (element.stats.kills + element.stats.assists) / deaths
            let tempData = {
                teamId: element.teamId,
                puuid: partMap[element.participantId].puuid,
                championId: element.championId,
                championIcon: buildRuntimeRiotAssetUrl(`/v1/champion-icons/${element.championId}.png`),
                spell1Icon: buildRuntimeRiotAssetUrl(`/DATA/Spells/Icons2D/${spellMap[element.spell1Id]}`),
                spell2Icon: buildRuntimeRiotAssetUrl(`/DATA/Spells/Icons2D/${spellMap[element.spell2Id]}`),
                kills: element.stats.kills,
                deaths: element.stats.deaths,
                assists: element.stats.assists,
                kda: parseFloat(kda.toFixed(1)),
                champLevel: element.stats.champLevel,
                gameName: partMap[element.participantId].gameName,
                item0: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item0]),
                item1: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item1]),
                item2: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item2]),
                item3: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item3]),
                item4: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item4]),
                item5: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item5]),
                item6: buildRuntimeGameItemIconUrl(itemIconMap[element.stats.item6]),
                goldEarned: element.stats.goldEarned,
                totalDamageDealtToChampions: element.stats.totalDamageDealtToChampions,// 伤害
                totalDamageTaken: element.stats.totalDamageTaken, // 承伤
                damageDealtToObjectives: element.stats.damageDealtToObjectives,// 资源
                visionScore: element.stats.visionScore, // 视野
                timeCCingOthers: element.stats.timeCCingOthers, // 控制
                pentaKills: element.stats.pentaKills > 0 ? 1 : 0,
                tier: 'unranked',
                rankText: '暂无排位记录',
            }
            if (element.teamId === 100) {
                tempPart1.push(tempData)
            } else {
                tempPart2.push(tempData)
            }
            allPuuid.push(tempData.puuid)
        })
        tempPart1.forEach(element => {
            teamTotal1.kills += element.kills
            teamTotal1.deaths += element.deaths
            teamTotal1.assists += element.assists
            teamTotal1.goldEarned += element.goldEarned
        })
        teamTotal1.data = tempPart1
        teamTotal1.maxData = computedMax(tempPart1)
        totalPart.push(teamTotal1)
        if (tempPart2.length > 0) {
            tempPart2.forEach(element => {
                teamTotal2.kills += element.kills
                teamTotal2.deaths += element.deaths
                teamTotal2.assists += element.assists
                teamTotal2.goldEarned += element.goldEarned
            })
            teamTotal2.data = tempPart2
            teamTotal2.maxData = computedMax(tempPart2)
            totalPart.push(teamTotal2)
        }
        allTotalMal.value = computedMax([...tempPart1, ...tempPart2])
        console.log('allTotalMal', allTotalMal.value)
        allPuuids.value = allPuuid
        teamInfo.value = totalPart
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
                    element.data.forEach(item => {
                        if (allPuuidRankedStatus[item.puuid].tier !== '' && allPuuidRankedStatus[item.puuid].queueType !== 'RANKED_TFT') {
                            let rankInfo = allPuuidRankedStatus[item.puuid]
                            item.rankText = queueTypeMap[rankInfo.queueType] + ' ' + rankMap[rankInfo.tier] + (rankInfo.division == 'NA' ? '' : rankInfo.division)
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
    fetchGameDetail(props.gameId)
});

watch(() => props.gameId, (newId, oldId) => {
    fetchGameDetail(newId)
})

const backgroundColor = (cur, target) => {
    return cur === target ? 'fade-green-background' : 'fade-red-background';
}
const emit = defineEmits(['checkout-puuid']);

const userClick = (puuid) => {
    emit('checkout-puuid', puuid);
};
const isCurrentUser = (itemPuuid) => {
    return itemPuuid === props.puuid;
};
const formatNumber = (num) => {
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    } else {
        return num.toString();
    }
}

const computedMax = (data) => {
    // 需要计算的属性
    let attr = new Set(['kills', 'deaths', 'assists', 'goldEarned', 'kda', 'totalDamageDealtToChampions', 'totalDamageTaken', 'visionScore', 'timeCCingOthers', 'damageDealtToObjectives']);
    let maxValues = data.reduce((acc, current) => {
        Object.keys(current).forEach(key => {
            if (attr.has(key) && !acc[key] || current[key] > acc[key]) {
                acc[key] = current[key];
            }
        });
        return acc;
    }, {});
    return { ...maxValues };
}
</script>
<style scoped>
.justift-center-container {
    display: grid;
    place-items: center;
}

.justift-left-element {
    justify-self: start;
}

.badge-level {
    font-size: 12px;
    height: 14px;
    line-height: 14px;
}

.fade-red-background {
    background: linear-gradient(to bottom, rgba(255, 71, 71, 0.2), rgba(236, 179, 179, 0));
}

.fade-green-background {
    background: linear-gradient(to bottom, rgba(69, 255, 100, 0.2), rgba(160, 255, 173, 0));
}

.fade-total-background {
    background-color: rgba(139, 186, 247, 0.2);
}

.text-overflow-container {
    width: 100%;
    /* 设置宽度 */
    white-space: nowrap;
    /* 禁止文本换行 */
    overflow: hidden;
    /* 隐藏溢出的内容 */
    text-overflow: ellipsis;
    /* 使用省略号表示溢出内容 */
}

.container {
    display: flex;
    justify-content: center;
    gap: 1px;
    margin: 0 auto;
    width: fit-content;
}

.square {
    width: 36px;
    height: 36px;
}

.small-items-container {
    width: 20px;
    height: 20px;
}

.victory {
    font-weight: bold;
    color: rgb(70, 182, 47);
    font-size: 16px;
}

.defeat {
    font-weight: bold;
    color: rgb(199, 33, 33);
    font-size: 16px;
}

.maxScore {
    color: rgb(74, 198, 236);
    font-weight: bold;
}
</style>
