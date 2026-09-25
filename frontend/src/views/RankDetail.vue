<template>
  <section class="match-detail" aria-label="对局详情" :aria-busy="loading">
    <a-skeleton
      v-if="loading && !data"
      active
      :paragraph="{ rows: 10 }"
      class="padded"
    />
    <div v-else-if="error" class="empty-state">
      <h3>暂时无法查看详情</h3>
      <p>
        {{
          offline && errorCode === 'CACHE_MISS'
            ? '本地只有摘要或详情已被清理。连接客户端并打开详情后可保存。'
            : error
        }}
      </p>
      <a-button v-if="!offline" @click="refresh">重试</a-button>
    </div>
    <template v-else-if="data">
      <header class="detail-header">
        <div>
          <h2>
            <span v-if="focusedTeam" :class="resultClass(focusedTeam.win)"
              >{{ focusedTeam.result }} · </span
            >{{ queueName }}
          </h2>
          <p class="detail-meta">
            <span class="muted"
              >{{ date(data.gameCreation) }} ·
              {{ duration(data.gameDuration) }} ·
              {{ data.gameVersion || '版本未知' }}</span
            ><span
              class="cache-label"
              :class="{ warning: !meta?.cached || meta?.refreshFailed }"
              >{{ sourceLabel
              }}<template v-if="meta?.fetchedAt"
                >（{{ date(meta.fetchedAt) }}）</template
              ></span
            >
          </p>
        </div>
        <div class="detail-actions">
          <span class="detail-legend muted">金色 = 全场最高</span
          ><a-button
            v-if="!offline"
            size="small"
            :loading="loading"
            @click="refresh"
            >重新获取</a-button
          >
        </div>
      </header>
      <section v-for="team in teams" :key="team.id" class="scoreboard">
        <header class="team-summary">
          <strong :class="resultClass(team.win)">{{ team.result }}</strong
          ><span>队伍 {{ team.id }}</span
          ><span>{{ value(team.kills) }} 击杀</span
          ><span>{{ number(team.gold) }} 金币</span>
        </header>
        <div
          v-for="player in team.players"
          :key="player.participantId"
          class="participant"
          :class="{ 'participant--focused': player.puuid === puuid }"
        >
          <div class="participant-identity">
            <span class="portrait-wrap">
              <AssetImage
                :src="asset(`/v1/champion-icons/${player.championId}.png`)"
                :label="`英雄 ${player.championId}`"
              />
              <span v-if="scoreOf(player).mvp" class="mvp-badge">MVP</span>
              <span v-else-if="scoreOf(player).svp" class="svp-badge"
                >SVP</span
              >
            </span>
            <div>
              <button
                class="text-button player-link"
                :disabled="!player.puuid"
                :title="player.name"
                @click="emit('checkout-puuid', player.puuid, player.name)"
              >
                <span
                  v-if="scoreOf(player).score !== null"
                  class="score-pill"
                  :class="`score-pill--${scoreTier(scoreOf(player).score)}`"
                  >{{ scoreOf(player).score.toFixed(1) }}</span
                ><span class="player-name">{{ player.name }}</span
                ><span v-if="player.puuid === puuid" class="mini-tag"
                  >当前</span
                ></button
              ><small
                class="muted"
                :title="`${rankLabel(ranks[player.puuid])} · 段位为最近查询快照，不代表参赛时段位`"
                >{{ rankLabel(ranks[player.puuid]) }}</small
              >
            </div>
          </div>
          <div
            class="participant-kda"
            :class="{ 'stat-top': isTop(player, 'kda') }"
          >
            <strong
              >{{ value(player.stats.kills) }}/{{
                value(player.stats.deaths)
              }}/{{ value(player.stats.assists) }}</strong
            ><small
              >KDA
              {{ player.kda === null ? '—' : player.kda.toFixed(1) }}</small
            >
          </div>
          <div class="loadout">
            <div class="spells">
              <AssetImage
                v-for="(spell, index) in [player.spell1Id, player.spell2Id]"
                :key="index"
                small
                :src="
                  spells[spell]
                    ? asset(`/DATA/Spells/Icons2D/${spells[spell]}`)
                    : ''
                "
                :label="`技能 ${spell || '—'}`"
              />
            </div>
            <div class="items">
              <template v-for="slot in 7" :key="slot">
                <AssetImage
                  v-if="player.stats[`item${slot - 1}`]"
                  small
                  :src="itemIcon(player.stats[`item${slot - 1}`])"
                  :label="`物品 ${player.stats[`item${slot - 1}`]}`"
                />
                <span v-else class="item-slot"></span>
              </template>
            </div>
          </div>
          <div
            class="participant-stat"
            :class="{
              'stat-top': isTop(player, 'totalDamageDealtToChampions'),
            }"
          >
            <small>伤害</small
            ><strong>{{
              number(player.stats.totalDamageDealtToChampions)
            }}</strong>
          </div>
          <div
            class="participant-stat"
            :class="{ 'stat-top': isTop(player, 'goldEarned') }"
          >
            <small>经济</small
            ><strong>{{ number(player.stats.goldEarned) }}</strong>
          </div>
          <div class="participant-highlights">
            <span :class="{ 'stat-top': isTop(player, 'totalDamageTaken') }"
              >承伤 {{ number(player.stats.totalDamageTaken) }}</span
            ><span
              :class="{
                'stat-top': isTop(player, 'damageDealtToObjectives'),
              }"
              >目标伤害 {{ number(player.stats.damageDealtToObjectives) }}</span
            ><span :class="{ 'stat-top': isTop(player, 'timeCCingOthers') }"
              >控制 {{ ccSeconds(player.stats.timeCCingOthers) }}秒</span
            ><span :class="{ 'stat-top': isTop(player, 'visionScore') }"
              >视野 {{ value(player.stats.visionScore) }}</span
            ><span
              v-if="(player.stats.pentaKills || 0) > 0"
              class="achievement-chip"
              >五杀</span
            ><span
              v-for="id in player.augments"
              :key="id"
              class="augment"
              :class="augmentRarity(id)"
              :title="augmentName(id)"
              ><AssetImage
                small
                :src="augmentIcon(id)"
                :label="augmentName(id)"
            /></span>
          </div>
        </div>
      </section>
    </template>
    <div v-else class="empty-state">
      <span class="empty-symbol">≡</span>
      <h3>选择一场对局</h3>
      <p>查看双方表现、装备和关键数据</p>
    </div>
  </section>
</template>
<script setup>
import { computed, watch } from 'vue'
import { useMatchDetail } from '@/composables/useMatchDetail'
import { buildRuntimeRiotAssetUrl as asset } from '@/utils/backend'
import { buildRuntimeGameItemIconUrl } from '@/utils/assets'
import { resolveQueueName } from '@/utils/queue'
import { computeMatchScores, scoreTier } from '@/utils/score'
import dicts from '@/model/dicts'
import AssetImage from './components/AssetImage.vue'
const props = defineProps({
  gameId: { default: 0 },
  scope: { default: '' },
  puuid: { default: '' },
  offline: Boolean,
})
const emit = defineEmits(['checkout-puuid', 'availability-change'])
const selection = computed(() => ({
  gameId: props.gameId,
  scope: props.scope,
  offline: props.offline,
}))
const { data, meta, ranks, loading, error, errorCode, refresh } =
  useMatchDetail(selection)
watch(meta, (value) => {
  if (value)
    emit('availability-change', {
      gameId: props.gameId,
      cached: !!value.cached,
    })
})
watch(errorCode, (code) => {
  if (props.offline && ['CACHE_MISS', 'CACHE_CORRUPT'].includes(code))
    emit('availability-change', { gameId: props.gameId, cached: false })
})
const spells = dicts.getFeDict('spell')
const items = dicts.getFeDict('gameItem')
const rankMap = dicts.getDict('rank')
const itemIcon = (id) =>
  id && items[id] ? buildRuntimeGameItemIconUrl(items[id]) : ''
const rankLabel = (rank) =>
  rank?.tier && rank.tier !== 'NONE'
    ? `最近段位：${rankMap[rank.tier] || rank.tier} ${rank.division === 'NA' ? '' : rank.division || ''}`
    : '暂无段位快照'
const queueName = computed(() =>
  resolveQueueName(
    dicts.getDict('queue'),
    data.value?.queueId,
    data.value?.gameMode,
  ),
)
const sourceLabel = computed(() =>
  meta.value?.refreshFailed
    ? '刷新失败，保留上次查看的数据'
    : meta.value?.source === 'disk-cache'
      ? '本地记录 · 可离线查看'
      : meta.value?.cached
        ? '刚从客户端获取 · 已保存，可离线查看'
        : '本次可查看，未保存到本地',
)
const date = (input) =>
  input ? new Date(input).toLocaleString('zh-CN') : '时间未知'
const duration = (seconds) =>
  seconds ? `${Math.floor(seconds / 60)}分${seconds % 60}秒` : '时长未知'
const value = (n) => (n == null ? '—' : n)
const number = (n) =>
  n == null ? '—' : n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n)
const players = computed(() => {
  const identities = new Map(
    (data.value?.participantIdentities || []).map((item) => [
      item.participantId,
      item.player || {},
    ]),
  )
  return (data.value?.participants || []).map((item) => {
    const identity = identities.get(item.participantId) || {}
    const stats = item.stats || {}
    return {
      ...item,
      stats,
      puuid: identity.puuid || '',
      name: identity.gameName
        ? `${identity.gameName}${identity.tagLine ? '#' + identity.tagLine : ''}`
        : identity.summonerName || '未知玩家',
      kda: [stats.kills, stats.deaths, stats.assists].every(Number.isFinite)
        ? Math.round(
            ((stats.kills + stats.assists) / Math.max(1, stats.deaths)) * 10,
          ) / 10
        : null,
      augments: [1, 2, 3, 4, 5, 6]
        .map((slot) => stats[`playerAugment${slot}`])
        .filter((id) => Number.isInteger(id) && id > 0),
    }
  })
})
// 全场最高值高亮维度：文字标记已移除，仅用颜色区分。
const leaderStats = [
  'kda',
  'goldEarned',
  'totalDamageDealtToChampions',
  'totalDamageTaken',
  'damageDealtToObjectives',
  'timeCCingOthers',
  'visionScore',
  'assists',
]
const statValue = (player, key) =>
  key === 'kda' ? player.kda : player.stats[key]
const maxStats = computed(() => {
  const result = {}
  leaderStats.forEach((key) => {
    const values = players.value
      .map((player) => statValue(player, key))
      .filter((value) => Number.isFinite(value))
    result[key] = values.length ? Math.max(...values) : null
  })
  return result
})
const isTop = (player, key) => {
  const value = statValue(player, key)
  return Number.isFinite(value) && value === maxStats.value[key]
}
// 海克斯大乱斗强化：LCU 对局详情 stats.playerAugment1~6 直接带出选择结果。
const augmentDict = dicts.getFeDict('augment')
const augmentOf = (id) => augmentDict[String(id)]
const augmentIcon = (id) => {
  const entry = augmentOf(id)
  return entry?.i ? asset(entry.i) : ''
}
const augmentName = (id) => augmentOf(id)?.n || `强化 ${id}`
const augmentRarity = (id) => {
  const rarity = augmentOf(id)?.r
  if (rarity === 'kPrismatic') return 'augment--prismatic'
  if (rarity === 'kGold') return 'augment--gold'
  if (rarity === 'kSilver') return 'augment--silver'
  return ''
}
const ccSeconds = (value) => (Number.isFinite(value) ? Math.round(value) : '—')
const resultClass = (win) =>
  win === null ? 'muted' : win ? 'win-text' : 'loss-text'
const sumStat = (members, key) => {
  const values = members.map((player) => player.stats[key])
  return values.every(Number.isFinite)
    ? values.reduce((sum, value) => sum + value, 0)
    : null
}
const teams = computed(() =>
  [...new Set(players.value.map((p) => p.teamId))].map((id) => {
    const members = players.value.filter((p) => p.teamId === id)
    const result = data.value.teams?.find((t) => t.teamId === id)?.win
    const win =
      result === 'Win' || result === true
        ? true
        : result === 'Fail' || result === false
          ? false
          : null
    return {
      id,
      players: members,
      win,
      result: win === null ? '结果未知' : win ? '胜利' : '失败',
      kills: sumStat(members, 'kills'),
      gold: sumStat(members, 'goldEarned'),
    }
  }),
)
const focusedTeam = computed(() =>
  teams.value.find((team) =>
    team.players.some((player) => player.puuid === props.puuid),
  ),
)
// 单局综合评分：胜方最高分标记 MVP，败方最高分标记 SVP。
const scores = computed(() =>
  computeMatchScores(
    players.value,
    teams.value.map((team) => ({ id: team.id, win: team.win })),
  ),
)
const scoreOf = (player) =>
  scores.value[player.participantId] || { score: null, mvp: false, svp: false }
</script>
