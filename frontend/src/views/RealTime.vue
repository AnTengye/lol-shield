<template>
  <div class="page realtime-page">
    <header class="page-heading">
      <div>
        <span class="eyebrow">LIVE COMPANION</span>
        <h1>实时对局</h1>
        <p class="muted">了解双方，专注这一场比赛。</p>
      </div>
      <div class="actions">
        <a-popover title="显示选项" trigger="click"
          ><template #content
            ><div class="preference-options">
              <a-checkbox
                :checked="preferences.showRank"
                @change="setPreference('showRank', $event.target.checked)"
                >段位</a-checkbox
              ><a-checkbox
                :checked="preferences.showParty"
                @change="setPreference('showParty', $event.target.checked)"
                >组队关系</a-checkbox
              >
            </div></template
          ><a-button>显示选项</a-button></a-popover
        ><a-button :disabled="!started" :loading="loading" @click="load"
          >刷新对局</a-button
        >
      </div>
    </header>
    <div v-if="!online || !started" class="empty-state panel waiting-state">
      <span class="empty-symbol">◇</span>
      <h2>{{ !online ? '等待连接英雄联盟客户端' : '等待对局开始' }}</h2>
      <p>
        {{
          !online
            ? '客户端未连接时，仍然可以查看已保存的战绩。'
            : '进入对局后，将在这里显示双方召唤师信息。'
        }}
      </p>
      <router-link to="/archive"><a-button>查看离线记录</a-button></router-link>
    </div>
    <a-skeleton
      v-else-if="loading && !game"
      active
      :paragraph="{ rows: 12 }"
      class="panel padded"
    />
    <a-alert v-if="error" type="warning" :message="error" show-icon />
    <template v-if="game && online && started">
      <a-alert
        v-if="game.partial"
        type="info"
        message="部分召唤师资料暂不可用，已展示当前取得的数据。"
        show-icon
      />
      <div class="match-overview panel">
        <span class="live-dot"></span><strong>{{ queueName }}</strong
        ><span class="muted">对局进行中</span
        ><span class="overview-time">更新于 {{ updatedAt }}</span>
      </div>
      <div class="match-board">
        <section v-for="(team, index) in teams" :key="team.id" class="team-row">
          <header class="team-header">
            <h2>
              {{ index === 0 ? '我方队伍' : '敌方队伍' }}
              <small>{{ team.users.length }} 位召唤师</small>
            </h2>
            <div v-if="preferences.showParty" class="party-legend">
              <span
                v-for="party in summarizeParties(team.users)"
                :key="party.partyKey"
                :style="{ color: party.partyColor }"
                >{{ party.partyLabel }} · {{ party.size }}人</span
              >
            </div>
          </header>
          <div class="player-grid">
            <article
              v-for="user in team.users"
              :key="user.puuid"
              class="player-card"
              :class="{ 'player-card--self': user.puuid === selfPuuid }"
            >
              <div class="player-portrait">
                <AssetImage
                  :src="
                    user.championId
                      ? asset(`/v1/champion-icons/${user.championId}.png`)
                      : ''
                  "
                  :label="user.name.gameName || '召唤师'"
                /><span v-if="user.puuid === selfPuuid" class="mini-tag"
                  >我</span
                >
              </div>
              <button
                class="player-nameplate"
                :title="
                  [user.name.gameName, user.name.tagLine]
                    .filter(Boolean)
                    .join('#')
                "
                @click.stop="openPlayerHistory(user)"
              >
                <strong>{{ user.name.gameName || '未知玩家' }}</strong
                ><small>#{{ user.name.tagLine || '—' }}</small>
              </button>
              <div v-if="preferences.showRank" class="rank-pill">
                {{ rankLabel(ranks[user.puuid]) }}
              </div>
              <div class="recent-performance">
                <strong :class="user.winRate >= 50 ? 'win-text' : 'muted'">{{
                  user.total ? `${user.winRate}%` : '—'
                }}</strong
                ><span>近 {{ user.total }} 场胜率</span>
              </div>
              <span
                v-if="preferences.showParty"
                class="party-badge"
                :style="{ color: user.partyColor }"
                >{{ user.partyLabel || '单排 / 未知' }}</span
              >
              <button
                class="history-action"
                @click.stop="openPlayerHistory(user)"
              >
                查看战绩 <span>↗</span>
              </button>
            </article>
          </div>
        </section>
      </div>
      <p class="muted source-note">
        近期战绩仅供参考 · 点击召唤师查看战绩与完整对局详情
      </p>
    </template>
    <PlayerHistoryPanel
      v-model:open="historyDrawerOpen"
      :player="selectedPlayer"
    />
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useStore } from 'vuex'
import { getGameRunning, getMulGameRankHighest } from '@/api/bog'
import { buildRuntimeRiotAssetUrl as asset } from '@/utils/backend'
import { buildPartyView, summarizeParties } from '@/utils/party'
import { resolveQueueName } from '@/utils/queue'
import dicts from '@/model/dicts'
import AssetImage from './components/AssetImage.vue'
import PlayerHistoryPanel from './components/PlayerHistoryPanel.vue'
const store = useStore()
const online = computed(() => store.getters['ws/getStatus'] === 1)
const started = computed(() => store.getters['ws/getGameStatus'] === 2)
const selfPuuid = computed(() => store.getters['ws/getUuid'])
const preferences = computed(() => store.state.ui.preferences)
const setPreference = (key, value) =>
  store.commit('ui/preference', { [key]: value })
const game = ref(null),
  loading = ref(false),
  error = ref(''),
  ranks = ref({}),
  updatedAt = ref('')
const historyDrawerOpen = ref(false),
  selectedPlayer = ref({})
let generation = 0,
  controller
const queueName = computed(() =>
  resolveQueueName(
    dicts.getDict('queue'),
    game.value?.queueId,
    game.value?.queueName,
  ),
)
const rankLabel = (rank) =>
  rank?.tier
    ? `${dicts.getDict('rank')[rank.tier] || rank.tier} ${rank.division === 'NA' ? '' : rank.division || ''}`
    : '暂无段位数据'
const teams = computed(() => {
  if (!game.value) return []
  const info = game.value
  const all = [
    ...(info.selfTeamInfo?.userList || []),
    ...(info.enemyTeamInfo?.userList || []),
  ]
  const parties = buildPartyView(info.preTeam || {}, all)
  return [info.selfTeamInfo, info.enemyTeamInfo]
    .filter(Boolean)
    .map((team) => ({
      id: team.teamId,
      users: (team.userList || []).map((user) => {
        const history = info.allGameHistory?.[user.puuid] || []
        return {
          ...user,
          name: info.userNameMap?.[user.puuid] || {},
          championId: info.skinMap?.[user.puuid]?.championId,
          total: history.length,
          winRate: history.length
            ? Math.round(
                (history.filter((g) => g.win).length * 100) / history.length,
              )
            : 0,
          ...parties.byPuuid[user.puuid],
        }
      }),
    }))
})
async function load() {
  const id = ++generation
  controller?.abort()
  controller = new AbortController()
  error.value = ''
  loading.value = false
  if (!online.value || !started.value) {
    game.value = null
    ranks.value = {}
    return
  }
  loading.value = true
  try {
    const result = await getGameRunning({ signal: controller.signal })
    if (id !== generation) return
    game.value = result.data
    ranks.value = {}
    updatedAt.value = new Date().toLocaleTimeString('zh-CN')
    const ids = teams.value.flatMap((team) =>
      team.users.map((user) => user.puuid),
    )
    if (ids.length)
      getMulGameRankHighest(ids, { signal: controller.signal })
        .then((result) => {
          if (id === generation)
            ranks.value = Object.fromEntries(
              result.data.map((row) => [row.puuid, row.data]),
            )
        })
        .catch(() => {})
  } catch (err) {
    if (id === generation && err.code !== 'ERR_CANCELED')
      error.value = err.message
  } finally {
    if (id === generation) loading.value = false
  }
}
watch([online, started], load, { immediate: true })
onBeforeUnmount(() => {
  generation++
  controller?.abort()
})
function openPlayerHistory(user) {
  selectedPlayer.value = {
    puuid: user.puuid,
    name: [user.name.gameName, user.name.tagLine].filter(Boolean).join('#'),
  }
  historyDrawerOpen.value = true
}
</script>
