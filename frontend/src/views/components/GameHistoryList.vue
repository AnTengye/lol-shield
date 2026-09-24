<template>
  <section class="history-list" aria-label="对局战绩">
    <div class="section-toolbar">
      <span
        >{{ offline ? '本地战绩' : '最近战绩' }}
        <small>{{ total }} 场</small></span
      ><a-button size="small" :loading="loading" @click="refresh"
        >刷新</a-button
      >
    </div>
    <p v-if="meta?.source === 'disk-cache'" class="muted source-note">
      本地记录<span v-if="meta.fetchedAt">
        · {{ formatDate(meta.fetchedAt) }}</span
      ><span v-if="meta.stale"> · 等待更新</span>
    </p>
    <p v-if="meta?.refreshFailed" class="warning-text source-note">
      刷新失败，保留上次查看的列表
    </p>
    <a-alert v-if="error" type="warning" :message="error" show-icon />
    <a-skeleton
      v-if="loading && !list.length"
      active
      :paragraph="{ rows: 8 }"
      class="padded"
    />
    <a-empty
      v-else-if="!list.length && !error"
      description="暂无已获取的对局"
    />
    <div
      v-else
      ref="scrollElement"
      class="history-scroll"
      @scroll="emit('scroll-change', $event.target.scrollTop)"
    >
      <button
        v-for="item in list"
        :key="`${item.scope}:${item.gameId}`"
        class="history-row"
        :class="{ selected: String(item.gameId) === String(selectedGameId) }"
        :aria-pressed="String(item.gameId) === String(selectedGameId)"
        @click="selectGame(item)"
      >
        <span
          class="result-stripe"
          :class="item.win ? 'win-bg' : 'loss-bg'"
        ></span>
        <AssetImage
          :src="
            buildRuntimeRiotAssetUrl(
              `/v1/champion-icons/${item.championId}.png`,
            )
          "
          :label="`英雄 ${item.championId}`"
        />
        <span class="history-copy"
          ><strong>{{
            resolveQueueName(queueMap, item.queueId, item.gameMode)
          }}</strong
          ><small
            >{{ formatDate(item.createTime) }} ·
            {{ duration(item.gameDuration) }}</small
          ><small>{{
            item.hasDetail ? '详情已保存' : '点击查看详情'
          }}</small></span
        >
        <span class="history-score"
          ><b :class="item.win ? 'win-text' : 'loss-text'">{{
            item.win ? '胜利' : '失败'
          }}</b
          ><span
            >{{ item.kills }}/{{ item.deaths }}/{{ item.assists }}</span
          ></span
        >
      </button>
    </div>
    <a-pagination
      v-if="showPagination && total > 20"
      :current="page"
      :total="total"
      :page-size="20"
      :show-size-changer="false"
      size="small"
      class="history-pagination"
      @change="emit('page-change', $event)"
    />
  </section>
</template>
<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { usePlayerHistory } from '@/composables/usePlayerHistory'
import { buildRuntimeRiotAssetUrl } from '@/utils/backend'
import { resolveQueueName } from '@/utils/queue'
import dicts from '@/model/dicts'
import AssetImage from './AssetImage.vue'
const props = defineProps({
  puuid: { default: '' },
  scope: { default: '' },
  page: { default: 1 },
  selectedGameId: { default: 0 },
  offline: Boolean,
  filters: { default: () => ({}) },
  autoSelectFirst: Boolean,
  showPagination: { default: true },
  scrollTop: { default: 0 },
})
const emit = defineEmits([
  'game-change',
  'page-loaded',
  'page-change',
  'scroll-change',
])
const query = computed(() => ({
  puuid: props.puuid,
  scope: props.scope,
  page: props.page,
  offline: props.offline,
  ...props.filters,
}))
const { list, total, meta, loading, error, refresh } = usePlayerHistory(query)
const queueMap = dicts.getDict('queue')
const scrollElement = ref(null)
const selectGame = (item) =>
  emit(
    'game-change',
    item.gameId,
    item.scope || meta.value?.scope || props.scope,
  )
const formatDate = (value) =>
  value
    ? new Date(value).toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '时间未知'
const duration = (value) => (value ? `${Math.floor(value / 60)}分` : '—')
watch(list, async (games) => {
  emit('page-loaded', { list: games, total: total.value, meta: meta.value })
  if (props.autoSelectFirst && games.length && !props.selectedGameId)
    selectGame(games[0])
  await nextTick()
  if (scrollElement.value) scrollElement.value.scrollTop = props.scrollTop
})
function markDetail({ gameId, cached }) {
  const item = list.value.find((item) => String(item.gameId) === String(gameId))
  if (item) item.hasDetail = cached
}
defineExpose({ fetchGameHistory: refresh, markDetail })
</script>
