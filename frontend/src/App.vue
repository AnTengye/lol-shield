<template>
  <a-config-provider :theme="themeConfig" :locale="zhCN">
    <contextHolder />
    <div class="app-shell">
      <aside class="sidebar">
        <router-link to="/running" class="brand"
          ><span class="brand-mark">S</span
          ><span class="nav-label"
            >LOL SHIELD<small>你的对局工作台</small></span
          ></router-link
        >
        <nav aria-label="主导航">
          <router-link
            v-for="item in navigation"
            :key="item.path"
            :to="item.path"
            :title="item.label"
            :aria-label="item.label"
            class="nav-link"
            ><component :is="item.icon" /><span class="nav-label">{{
              item.label
            }}</span
            ><span
              v-if="item.path === '/running' && started"
              class="nav-live-dot"
            ></span
          ></router-link>
        </nav>
        <div class="sidebar-bottom">
          <router-link to="/settings" class="nav-link" title="设置"
            ><SettingOutlined /><span class="nav-label">设置</span></router-link
          >
          <div class="account-summary">
            <AssetImage
              :src="
                user.profileIconId
                  ? asset(`/v1/profile-icons/${user.profileIconId}.jpg`)
                  : ''
              "
              label="账号"
              small
            />
            <div class="nav-label">
              <strong>{{ user.gameName || '未连接账号' }}</strong
              ><small>{{
                user.tagLine ? '#' + user.tagLine : '本地记录随时可用'
              }}</small>
            </div>
          </div>
        </div>
      </aside>
      <div class="main-shell">
        <header class="topbar">
          <span class="muted">{{ route.meta.title || 'LOL Shield' }}</span>
          <div class="connection-status">
            <span :class="backendOnline ? 'status-good' : 'status-waiting'"
              >● {{ backendOnline ? '本地服务正常' : '本地服务连接中' }}</span
            ><span :class="online ? 'status-good' : 'muted'">{{
              online ? '客户端已连接' : '客户端未连接'
            }}</span>
          </div>
        </header>
        <div v-if="gameNotice && route.path !== '/running'" class="game-notice">
          <span>对局已开始，可查看双方召唤师信息。</span
          ><a-button
            size="small"
            type="primary"
            @click="showRunning"
            >查看实时对局</a-button
          ><button
            aria-label="关闭开局提示"
            class="text-button"
            @click="gameNotice = false"
          >
            ×
          </button>
        </div>
        <UpdateNotice v-if="route.path !== '/settings'" />
        <main class="main-content"><router-view /></main>
      </div>
    </div>
  </a-config-provider>
</template>
<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Modal, theme } from 'ant-design-vue'
import { getCurrentWindow } from '@tauri-apps/api/window'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import {
  ThunderboltOutlined,
  HistoryOutlined,
  DatabaseOutlined,
  SettingOutlined,
} from '@ant-design/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { getStatus, getUser } from '@/api/bog'
import { createWebSocket, destroyWebSocket } from '@/websocket'
import { buildRuntimeRiotAssetUrl as asset } from '@/utils/backend'
import AssetImage from '@/views/components/AssetImage.vue'
import UpdateNotice from '@/views/components/UpdateNotice.vue'
const route = useRoute(),
  router = useRouter(),
  store = useStore()
function showRunning() { router.push('/running'); gameNotice.value = false }
const user = ref({}),
  gameNotice = ref(false)
const [modal, contextHolder] = Modal.useModal()
let unlistenClose,
  closeConfirmation = false
const online = computed(() => store.getters['ws/getStatus'] === 1)
const started = computed(() => store.getters['ws/getGameStatus'] === 2)
const backendOnline = computed(() => store.state.ui.backendOnline)
const navigation = [
  { path: '/running', label: '实时对局', icon: ThunderboltOutlined },
  { path: '/rank', label: '战绩中心', icon: HistoryOutlined },
  { path: '/archive', label: '离线记录', icon: DatabaseOutlined },
]
const themeConfig = {
  algorithm: theme.darkAlgorithm,
  token: {
    colorPrimary: '#4DABF7',
    colorBgBase: '#0B1018',
    colorBgContainer: '#121B27',
    colorText: '#E8EEF7',
    colorTextSecondary: '#9CAEC4',
    colorBorder: '#29364A',
    borderRadius: 8,
    fontSize: 14,
  },
}
let timer,
  stopped = false,
  polling = false
async function pollStatus() {
  if (polling) return
  polling = true
  try {
    const response = await getStatus()
    if (stopped) return
    store.commit('ui/backendOnline', true)
    store.commit('ws/setWsRes', response.data)
  } catch {
    if (!stopped) {
      store.commit('ui/backendOnline', false)
      store.commit('ws/reset')
    }
  } finally {
    polling = false
  }
}
watch(
  () => [online.value, store.getters['ws/getUuid']],
  async ([connected], _, onCleanup) => {
    let active = true
    onCleanup(() => {
      active = false
    })
    if (!connected) {
      user.value = {}
      return
    }
    try {
      const result = await getUser()
      if (active) user.value = result.data
    } catch {
      /* 用户资料失败不影响离线入口 */
    }
  },
)
watch(started, (value) => {
  if (value) {
    if (
      store.state.ui.preferences.autoNavigate &&
      !store.state.ui.detailOpen &&
      !store.state.ui.settingsDirty
    )
      router.push('/running')
    else gameNotice.value = true
  } else gameNotice.value = false
})
onMounted(async () => {
  createWebSocket(store)
  pollStatus()
  timer = setInterval(pollStatus, 5000)
  if (store.state.ui.preferences.autoCheckUpdate)
    store.dispatch('ui/checkUpdate', { silent: true })
  if (window.__TAURI_INTERNALS__) {
    const desktop = getCurrentWindow()
    const unlisten = await desktop.onCloseRequested((event) => {
      if (!store.state.ui.settingsDirty) return
      event.preventDefault()
      if (closeConfirmation) return
      closeConfirmation = true
      modal.confirm({
        title: '尚有未保存的配置',
        content: '关闭窗口将丢弃本次更改。',
        okText: '丢弃并关闭',
        cancelText: '继续编辑',
        onOk: () => desktop.destroy(),
        afterClose: () => {
          closeConfirmation = false
        },
      })
    })
    if (stopped) unlisten()
    else unlistenClose = unlisten
  }
})
onBeforeUnmount(() => {
  stopped = true
  clearInterval(timer)
  unlistenClose?.()
  destroyWebSocket()
})
</script>
