<template>
  <div
    class="history-workspace"
    :class="{ 'has-detail': current.gameId, compact }"
  >
    <header class="workspace-header">
      <div class="workspace-title">
        <button
          v-if="stack.length > 1"
          class="text-button"
          @click="stack.pop()"
        >
          ← 返回上一位玩家
        </button>
        <h3>
          <span class="workspace-name">{{ current.name || '召唤师战绩' }}</span
          ><a-popover trigger="hover" placement="bottomLeft">
            <template #content>
              <div class="profile-popover">
                <p>
                  <span class="muted">大区</span>
                  {{ current.scope || '连接后识别大区' }}
                </p>
                <p>
                  <span class="muted">等级</span>
                  {{ profile?.summonerLevel || '—' }}
                </p>
                <p>
                  <span class="muted">最近段位</span>
                  {{ rankSnapshot ? rankLabel : '暂无段位快照' }}
                </p>
                <p v-if="profileMeta?.fetchedAt">
                  <span class="muted">资料更新</span>
                  {{ new Date(profileMeta.fetchedAt).toLocaleString('zh-CN') }}
                </p>
              </div>
            </template>
            <button
              class="profile-trigger"
              :title="profileTitle"
              aria-label="查看召唤师资料"
            >
              ⓘ
            </button>
          </a-popover>
        </h3>
      </div>
      <div class="actions">
        <a-button v-if="current.gameId" size="small" @click="closeDetail"
          >返回战绩列表</a-button
        ><slot name="actions" :current="current" />
      </div>
    </header>
    <div class="workspace-columns">
      <GameHistoryList
        ref="historyList"
        v-show="!compact || !current.gameId"
        :key="`${current.puuid}:${cacheEpoch}`"
        :puuid="current.puuid"
        :scope="current.scope"
        :page="current.page"
        :scroll-top="current.scrollTop"
        :selected-game-id="current.gameId || current.lastGameId"
        :offline="offline"
        :filters="filters"
        :auto-select-first="autoSelectFirst"
        @game-change="select"
        @page-loaded="pageLoaded"
        @page-change="
          (page) => {
            current.page = page
            current.scrollTop = 0
          }
        "
        @scroll-change="current.scrollTop = $event"
      />
      <RankDetail
        v-if="!compact || current.gameId"
        :key="cacheEpoch"
        :game-id="current.gameId"
        :scope="current.scope"
        :puuid="current.puuid"
        :offline="offline"
        @checkout-puuid="visit"
        @availability-change="historyList?.markDetail($event)"
      />
    </div>
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useStore } from 'vuex'
import { getPlayer, getGameRankHighest } from '@/api/bog'
import dicts from '@/model/dicts'
import GameHistoryList from './GameHistoryList.vue'
import RankDetail from '../RankDetail.vue'
const props = defineProps({
  player: { default: () => ({}) },
  offline: Boolean,
  compact: Boolean,
  autoSelectFirst: Boolean,
  filters: { default: () => ({}) },
  stateKey: { default: '' },
})
const emit = defineEmits(['selection-change'])
const store = useStore()
const stack = ref([])
const profile = ref(null),
  profileMeta = ref(null),
  rankSnapshot = ref(null)
const rankLabel = computed(() =>
  rankSnapshot.value?.tier && rankSnapshot.value.tier !== 'NONE'
    ? `${dicts.getDict('rank')[rankSnapshot.value.tier] || rankSnapshot.value.tier} ${rankSnapshot.value.division === 'NA' ? '' : rankSnapshot.value.division || ''}`
    : '暂无排位',
)
// 资料概况仅供悬浮提示使用，头部不再占用文案行。
const historyList = ref(null)
const cacheEpoch = computed(() => store.state.ui.cacheGeneration)
const current = computed(() => stack.value[stack.value.length - 1] || {})
const profileTitle = computed(() => {
  const parts = []
  if (current.value.scope) parts.push(current.value.scope)
  if (profile.value) parts.push(`等级 ${profile.value.summonerLevel || '—'}`)
  if (rankSnapshot.value) parts.push(`最近段位 ${rankLabel.value}`)
  if (profileMeta.value?.fetchedAt)
    parts.push(
      `资料更新于 ${new Date(profileMeta.value.fetchedAt).toLocaleString('zh-CN')}`,
    )
  return parts.join(' · ') || '暂无资料'
})
function entry(player) {
  return {
    puuid: player.puuid || '',
    name: player.name || '',
    scope: player.scope || '',
    gameId: player.gameId || 0,
    lastGameId: 0,
    page: 1,
    scrollTop: 0,
  }
}
watch(
  () => props.player,
  (player) => {
    const snapshot = props.stateKey && store.state.ui.snapshots[props.stateKey]
    stack.value =
      snapshot?.[0]?.puuid === player.puuid &&
      (!player.scope || snapshot?.[0]?.scope === player.scope) &&
      (!player.gameId ||
        String(snapshot.at(-1)?.gameId) === String(player.gameId))
        ? JSON.parse(JSON.stringify(snapshot))
        : [entry(player)]
  },
  { immediate: true, deep: true },
)
watch(
  () => [current.value.puuid, current.value.scope],
  async ([puuid, scope], _, onCleanup) => {
    profile.value = null
    profileMeta.value = null
    rankSnapshot.value = null
    if (!puuid || !scope) return
    const controller = new AbortController()
    onCleanup(() => controller.abort())
    const options = {
      scope,
      policy: props.offline ? 'cache-only' : 'cache-first',
      signal: controller.signal,
    }
    getGameRankHighest(puuid, options)
      .then((result) => {
        if (!controller.signal.aborted) rankSnapshot.value = result.data
      })
      .catch(() => {})
    try {
      const result = await getPlayer(puuid, options)
      if (!controller.signal.aborted) {
        profile.value = result.data
        profileMeta.value = result.meta
        if (!current.value.name)
          current.value.name = [result.data.gameName, result.data.tagLine]
            .filter(Boolean)
            .join('#')
      }
    } catch {
      /* 资料缺失不阻断战绩 */
    }
  },
  { immediate: true },
)
watch(
  stack,
  (value) => {
    emit('selection-change', current.value)
    store.commit('ui/detailOpen', !!current.value.gameId)
    if (props.stateKey)
      store.commit('ui/snapshot', {
        key: props.stateKey,
        value: JSON.parse(JSON.stringify(value)),
      })
  },
  { deep: true, immediate: true },
)
watch(
  () => props.filters,
  () => {
    current.value.page = 1
    current.value.scrollTop = 0
    current.value.gameId = 0
    current.value.lastGameId = 0
  },
  { deep: true },
)
watch(cacheEpoch, () => {
  stack.value = [entry({})]
  profile.value = null
})
function pageLoaded({ list, meta }) {
  if (meta?.scope && !current.value.scope) current.value.scope = meta.scope
  if (
    meta &&
    current.value.lastGameId &&
    !list.some(
      (item) => String(item.gameId) === String(current.value.lastGameId),
    )
  )
    current.value.lastGameId = 0
}
function select(gameId, scope) {
  current.value.gameId = gameId
  if (scope) current.value.scope = scope
}
function visit(puuid, name) {
  if (puuid && puuid !== current.value.puuid)
    stack.value.push(entry({ puuid, name, scope: current.value.scope }))
}
function closeDetail() {
  current.value.lastGameId = current.value.gameId
  current.value.gameId = 0
}
function back() {
  if (current.value.gameId) {
    closeDetail()
    return true
  }
  if (stack.value.length > 1) {
    stack.value.pop()
    return true
  }
  return false
}
onBeforeUnmount(() => store.commit('ui/detailOpen', false))
defineExpose({ back, current })
</script>
