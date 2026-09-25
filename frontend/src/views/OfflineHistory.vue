<template>
  <div class="page history-page">
    <header class="page-heading">
      <div>
        <h1>离线记录</h1>
      </div>
      <a-button @click="load">刷新本地列表</a-button>
    </header>
    <a-alert v-if="error" type="warning" :message="error" show-icon />
    <div class="archive-controls panel">
      <a-select
        show-search
        :filter-option="false"
        placeholder="搜索本地召唤师"
        :value="selectedKey"
        style="min-width: 240px"
        @search="searchPlayers"
        @change="choose"
        ><a-select-option v-for="p in players" :key="`${p.scope}|${p.puuid}`"
          >{{ p.gameName || p.puuid }}#{{ p.tagLine }} ·
          {{ p.scope }}</a-select-option
        ></a-select
      ><a-select
        v-model:value="filters.win"
        allow-clear
        placeholder="胜负"
        style="width: 100px"
        ><a-select-option value="true">胜利</a-select-option
        ><a-select-option value="false">失败</a-select-option></a-select
      ><a-select
        v-model:value="filters.queueId"
        allow-clear
        show-search
        option-filter-prop="label"
        placeholder="游戏模式"
        :options="queues"
        style="width: 180px"
      /><a-range-picker @change="changeDates" /><a-checkbox
        v-model:checked="filters.detailsOnly"
        >仅已保存详情</a-checkbox
      >
    </div>
    <div
      v-if="selected"
      class="archive-count muted"
      :title="`最后查看 ${new Date(selected.lastViewed).toLocaleString('zh-CN')}`"
    >
      本地保存 {{ selected.summaries }} 场 · {{ selected.details }} 场可查看详情
    </div>
    <HistoryWorkspace
      v-if="selected"
      :player="player"
      :filters="filters"
      :compact="width < 1200"
      offline
      state-key="archive"
    />
    <div v-else class="empty-state panel">
      <span class="empty-symbol">▤</span>
      <h2>{{ loading ? '正在读取本地记录' : '选择一位已保存的召唤师' }}</h2>
      <p>
        浏览过的列表会保存摘要；打开过的详情可以离线查看。<br />缓存达到 500 MB
        后将自动清理较久未访问的记录。
      </p>
      <router-link to="/rank"
        ><a-button type="primary">前往战绩中心</a-button></router-link
      >
    </div>
    <a-pagination
      v-if="total > 100"
      size="small"
      :current="page + 1"
      :total="total"
      :page-size="100"
      :show-size-changer="false"
      @change="
        (value) => {
          page = value - 1
          load()
        }
      "
    />
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useStore } from 'vuex'
import { getArchivePlayers } from '@/api/bog'
import { useWindowWidth } from '@/composables/useWindowWidth'
import dicts from '@/model/dicts'
import HistoryWorkspace from './components/HistoryWorkspace.vue'
const store = useStore(),
  width = useWindowWidth()
const players = ref([]),
  selected = ref(null),
  loading = ref(false),
  error = ref(''),
  total = ref(0),
  page = ref(0),
  search = ref('')
const filters = reactive({
  win: undefined,
  queueId: undefined,
  detailsOnly: false,
  from: undefined,
  to: undefined,
})
const queues = Object.entries(dicts.getDict('queue'))
  .filter(([id]) => /^\d+$/.test(id))
  .map(([value, label]) => ({ value, label }))
const selectedKey = computed(() =>
  selected.value
    ? `${selected.value.scope}|${selected.value.puuid}`
    : undefined,
)
const player = computed(() =>
  selected.value
    ? {
        puuid: selected.value.puuid,
        scope: selected.value.scope,
        name: `${selected.value.gameName}#${selected.value.tagLine}`,
      }
    : {},
)
let request = 0,
  timer
async function load() {
  const id = ++request,
    active = selected.value || store.state.ui.snapshots.archive?.[0]
  const key = selectedKey.value
  loading.value = true
  error.value = ''
  try {
    const [result, current] = await Promise.all([
      getArchivePlayers({
        search: search.value,
        page: page.value,
        pageSize: 100,
      }),
      active?.puuid
        ? getArchivePlayers({
            scope: active.scope,
            puuid: active.puuid,
            pageSize: 1,
          })
        : null,
    ])
    if (id !== request) return
    players.value = result.data.list
    total.value = result.data.total
    if (current && selectedKey.value === key)
      selected.value = current.data.list[0] || null
  } catch (err) {
    if (id === request) error.value = err.message
  } finally {
    if (id === request) loading.value = false
  }
}
function searchPlayers(value) {
  search.value = value
  page.value = 0
  clearTimeout(timer)
  timer = setTimeout(load, 250)
}
function choose(key) {
  selected.value = players.value.find((p) => `${p.scope}|${p.puuid}` === key)
}
function changeDates(values) {
  filters.from = values?.[0]?.startOf('day').valueOf()
  filters.to = values?.[1]?.endOf('day').valueOf()
}
watch(
  () => store.state.ui.cacheGeneration,
  () => {
    selected.value = null
    players.value = []
    load()
  },
)
onMounted(load)
onBeforeUnmount(() => {
  request++
  clearTimeout(timer)
})
</script>
