<template>
  <div class="page history-page">
    <header class="page-heading">
      <div>
        <span class="eyebrow">MATCH HISTORY</span>
        <h1>战绩中心</h1>
        <p class="muted">每一次交锋，都值得回顾。</p>
      </div>
      <div class="actions">
        <router-link v-if="route.query.from === '/running'" to="/running"
          ><a-button>返回实时对局</a-button></router-link
        ><a-select
          placeholder="最近查看玩家"
          :value="undefined"
          style="width: 200px"
          @dropdown-visible-change="loadRecent"
          @change="selectRecent"
          ><a-select-option v-for="p in recent" :key="`${p.scope}|${p.puuid}`"
            >{{ p.gameName || p.puuid }}#{{ p.tagLine }} ·
            {{ p.scope }}</a-select-option
          ></a-select
        ><a-button :disabled="!self" @click="mine">我的战绩</a-button>
      </div>
    </header>
    <a-alert
      v-if="!online"
      type="info"
      show-icon
      message="客户端未连接；您可以在离线记录中查看已经保存的对局。"
    />
    <HistoryWorkspace
      v-if="player.puuid"
      :key="resetKey"
      :player="player"
      :compact="width < 1100"
      auto-select-first
      state-key="rank"
    />
    <div v-else class="empty-state panel">
      <h2>尚未选择召唤师</h2>
      <p>连接客户端查看我的战绩，或浏览已有离线记录。</p>
      <router-link to="/archive"
        ><a-button type="primary">打开离线记录</a-button></router-link
      >
    </div>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { getArchivePlayers } from '@/api/bog'
import { useWindowWidth } from '@/composables/useWindowWidth'
import HistoryWorkspace from './components/HistoryWorkspace.vue'
const route = useRoute(),
  router = useRouter(),
  store = useStore(),
  width = useWindowWidth()
const recent = ref([]),
  resetKey = ref(0)
const self = computed(() => store.getters['ws/getUuid'])
const online = computed(() => store.getters['ws/getStatus'] === 1)
const player = computed(() => ({
  puuid: route.query.puuid || self.value || '',
  scope: route.query.scope || '',
  name: route.query.name || (route.query.puuid ? '' : '我的战绩'),
  gameId: Number(route.query.gameId) || 0,
}))
async function loadRecent(open = true) {
  if (open)
    try {
      recent.value = (await getArchivePlayers({ pageSize: 100 })).data.list
    } catch {
      recent.value = []
    }
}
function selectRecent(key) {
  const p = recent.value.find((p) => `${p.scope}|${p.puuid}` === key)
  if (p)
    router.replace({
      path: '/rank',
      query: {
        puuid: p.puuid,
        scope: p.scope,
        name: `${p.gameName}#${p.tagLine}`,
      },
    })
}
async function mine() {
  store.commit('ui/snapshot', { key: 'rank', value: [] })
  await router.replace('/rank')
  resetKey.value++
}
onMounted(loadRecent)
</script>
