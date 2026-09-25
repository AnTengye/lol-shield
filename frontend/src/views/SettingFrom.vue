<template>
  <div class="page settings-page">
    <header class="page-heading">
      <div>
        <h1>设置</h1>
      </div>
    </header>
    <contextHolder />
    <a-alert v-if="error" type="warning" :message="error" show-icon />
    <section class="panel settings-section">
      <header>
        <h2>对局自动化</h2>
        <p class="muted">仅在客户端对应阶段生效，修改后请保存。</p>
      </header>
      <div class="setting-row">
        <div>
          <strong>自动接受对局</strong>
          <p>找到比赛时自动确认准备就绪。</p>
        </div>
        <a-switch v-model:checked="form.autoConfirm" :disabled="!loaded" />
      </div>
      <div class="setting-row">
        <div>
          <strong>自动选择英雄</strong>
          <p>轮到选择英雄时尝试选用；留空表示关闭。</p>
        </div>
        <a-select
          v-model:value="form.autoPick"
          allow-clear
          show-search
          option-filter-prop="label"
          :options="champions"
          placeholder="不自动选用"
          style="width: 240px"
          :disabled="!loaded"
        />
      </div>
      <div class="setting-row">
        <div>
          <strong>自动禁用英雄</strong>
          <p>进入禁用阶段时尝试禁用指定英雄。</p>
        </div>
        <a-select
          v-model:value="form.autoBan"
          allow-clear
          show-search
          option-filter-prop="label"
          :options="champions"
          placeholder="不自动禁用"
          style="width: 240px"
          :disabled="!loaded"
        />
      </div>
      <div class="setting-footer">
        <span class="muted">{{
          dirty ? '有尚未保存的更改' : '配置已同步'
        }}</span
        ><a-button
          type="primary"
          :loading="saving"
          :disabled="!loaded || !dirty"
          @click="save"
          >保存更改</a-button
        >
      </div>
    </section>
    <section class="panel settings-section">
      <header>
        <h2>显示与交互</h2>
        <p class="muted">这些偏好即时保存到本机。</p>
      </header>
      <div class="setting-row">
        <div>
          <strong>开局自动切换</strong>
          <p>查看详情或编辑配置期间仍只提示，不打断当前操作。</p>
        </div>
        <a-switch
          :checked="preferences.autoNavigate"
          @change="preference('autoNavigate', $event)"
        />
      </div>
      <div class="setting-row">
        <div>
          <strong>实时对局显示</strong>
          <p>自定义召唤师卡片的信息。</p>
        </div>
        <div>
          <a-checkbox
            :checked="preferences.showRank"
            @change="preference('showRank', $event.target.checked)"
            >段位</a-checkbox
          ><a-checkbox
            :checked="preferences.showParty"
            @change="preference('showParty', $event.target.checked)"
            >组队关系</a-checkbox
          >
        </div>
      </div>
    </section>
    <section class="panel settings-section">
      <header>
        <h2>本地数据</h2>
        <p class="muted">
          浏览即保存，较久未访问的数据会自动淘汰。容量上限 500 MB。
        </p>
      </header>
      <a-alert
        v-if="stats && !stats.healthy"
        type="warning"
        :message="stats.message || '本地缓存不可用'"
      />
      <div class="storage-summary">
        <strong>{{ mb(stats?.usedBytes) }} <small>/ 500 MB</small></strong
        ><a-button size="small" @click="loadStats">刷新容量</a-button>
      </div>
      <a-progress
        :percent="Math.min(100, (stats?.usedBytes || 0) / 5000000)"
        :show-info="false"
      />
      <p class="muted">
        {{ stats?.players || 0 }} 位玩家 · {{ stats?.summaries || 0 }} 条摘要 ·
        {{ stats?.details || 0 }} 场详情 · {{ stats?.assets || 0 }} 张图片
      </p>
      <div class="setting-footer">
        <span class="muted">{{
          clearResult || '清理缓存不会修改自动化配置。'
        }}</span>
        <div class="actions">
          <a-button :disabled="!stats?.healthy" @click="confirmClear('assets')"
            >清理图片缓存</a-button
          ><a-button
            danger
            :disabled="!stats?.healthy"
            @click="confirmClear('all')"
            >清空战绩缓存</a-button
          >
        </div>
      </div>
    </section>
    <section class="panel settings-section">
      <header>
        <h2>版本与更新</h2>
        <p class="muted">桌面版打开时在后台静默检查，不打断当前操作。</p>
      </header>
      <div class="setting-row">
        <div>
          <strong>当前版本</strong>
          <p>{{ versionLabel }}</p>
        </div>
        <a-button :loading="checking" :disabled="updateBusy" @click="checkNow"
          >检查更新</a-button
        >
      </div>
      <div class="setting-row">
        <div>
          <strong>启动时自动检查</strong>
          <p>发现新版本时只提示，不会自动下载或安装。</p>
        </div>
        <a-switch
          :checked="preferences.autoCheckUpdate"
          @change="preference('autoCheckUpdate', $event)"
        />
      </div>
      <div v-if="update" class="setting-row">
        <div>
          <strong>发现新版本 v{{ update.version }}</strong>
          <p>{{ updateHint }}</p>
        </div>
        <a-button
          type="primary"
          :loading="updateBusy"
          :disabled="updateBusy"
          @click="installNow"
          >一键更新</a-button
        >
      </div>
      <a-progress
        v-if="updateBusy"
        :percent="updateProgress"
        :show-info="false"
      />
      <a-alert
        v-if="updateError"
        type="warning"
        :message="updateError"
        show-icon
      />
    </section>
    <footer class="about muted">
      LOL Shield {{ desktopVersion || version }} · 本地运行，专注对局
    </footer>
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { useStore } from 'vuex'
import { message, Modal } from 'ant-design-vue'
import {
  getConfig,
  updateConfig,
  getCacheStats,
  clearCache,
  getVersion,
} from '@/api/bog'
const store = useStore(),
  preferences = computed(() => store.state.ui.preferences)
const update = computed(() => store.state.ui.update),
  updateBusy = computed(() => store.getters['ui/updateBusy']),
  updateProgress = computed(() => store.state.ui.updateProgress),
  updateError = computed(() => store.state.ui.updateError),
  desktopVersion = computed(() => store.state.ui.appVersion),
  checking = computed(() => store.state.ui.updateStatus === 'checking')
const statusText = computed(() => {
  const state = store.state.ui
  if (state.updateStatus === 'checking') return '正在检查更新…'
  if (state.updateStatus === 'downloading')
    return state.updateDetail || '正在下载更新…'
  if (state.updateStatus === 'installing')
    return state.updateDetail || '正在安装更新…'
  if (state.updateStatus === 'failed') return state.updateError || '更新失败'
  if (state.update) return `发现新版本 v${state.update.version}`
  if (state.updateCheckedAt) return '已是最新版本'
  return '尚未检查更新'
})
const versionLabel = computed(() =>
  [desktopVersion.value || version.value || '未知', statusText.value]
    .filter(Boolean)
    .join(' · '),
)
const updateHint = computed(
  () =>
    store.state.ui.updateDetail || '下载后自动安装并重启，无需手动操作。',
)
async function checkNow() {
  const result = await store.dispatch('ui/checkUpdate', { silent: false })
  if (result) message.success(`发现新版本 v${result.version}`)
  else if (!store.state.ui.updateError) message.success('已是最新版本')
}
const installNow = () => store.dispatch('ui/installUpdate')
const [modal, contextHolder] = Modal.useModal()
const form = reactive({
  autoConfirm: false,
  autoPick: undefined,
  autoBan: undefined,
})
const baseline = ref(''),
  loaded = ref(false),
  saving = ref(false),
  error = ref(''),
  champions = ref([]),
  stats = ref(null),
  clearResult = ref(''),
  version = ref('')
const serialize = () =>
  JSON.stringify({
    autoConfirm: form.autoConfirm,
    autoPick: form.autoPick || 0,
    autoBan: form.autoBan || 0,
  })
const dirty = computed(() => loaded.value && serialize() !== baseline.value)
watch(dirty, (value) => store.commit('ui/settingsDirty', value))
const preference = (key, value) =>
  store.commit('ui/preference', { [key]: value })
const mb = (bytes) => `${((bytes || 0) / 1000000).toFixed(1)} MB`
async function loadStats() {
  try {
    stats.value = (await getCacheStats()).data
  } catch (err) {
    error.value = err.message
  }
}
async function save() {
  if (saving.value) return
  const submitted = serialize()
  const values = JSON.parse(submitted)
  saving.value = true
  try {
    await updateConfig(values.autoConfirm, values.autoPick, values.autoBan)
    baseline.value = submitted
    message.success('配置已保存')
  } catch (err) {
    message.error(err.message)
  } finally {
    saving.value = false
  }
}
function confirmClear(kind) {
  modal.confirm({
    title: kind === 'all' ? '清空全部本地战绩？' : '清理缓存图片？',
    content:
      kind === 'all'
        ? '已保存的玩家、摘要和详情将被移除，离线时无法恢复。自动化配置不受影响。'
        : '数值战绩会保留，缺少的图标将在下次连接客户端查看时重新获取。',
    okText: '确认清理',
    cancelText: '取消',
    okButtonProps: { danger: true },
    async onOk() {
      try {
        const response = await clearCache(kind)
        stats.value = response.data.stats
        clearResult.value = `刚刚释放 ${mb(response.data.freedBytes)}`
        if (kind === 'all') store.commit('ui/cacheCleared')
        message.success('清理完成')
      } catch (err) {
        message.error(err.message)
        throw err
      }
    },
  })
}
onBeforeRouteLeave(
  () =>
    !dirty.value ||
    new Promise((resolve) =>
      modal.confirm({
        title: '尚有未保存的配置',
        content: '离开将丢弃本次更改。',
        okText: '丢弃并离开',
        cancelText: '继续编辑',
        onOk: () => resolve(true),
        onCancel: () => resolve(false),
      }),
    ),
)
function beforeUnload(event) {
  if (dirty.value) {
    event.preventDefault()
    event.returnValue = ''
  }
}
onMounted(async () => {
  window.addEventListener('beforeunload', beforeUnload)
  loadStats()
  store.dispatch('ui/loadAppVersion')
  getVersion()
    .then((result) => {
      version.value = result.data.version
    })
    .catch(() => {})
  fetch('/champion.json')
    .then((result) => result.json())
    .then((result) => {
      champions.value = Object.values(result.data).map((champion) => ({
        value: Number(champion.key),
        label: `${champion.title} · ${champion.name}`,
      }))
    })
    .catch(() => {})
  try {
    const result = (await getConfig()).data.game
    Object.assign(form, {
      autoConfirm: result.auto_confirm,
      autoPick: result.auto_pick || undefined,
      autoBan: result.auto_ban || undefined,
    })
    baseline.value = serialize()
    loaded.value = true
  } catch (err) {
    error.value = err.message
  }
})
onBeforeUnmount(() => {
  store.commit('ui/settingsDirty', false)
  window.removeEventListener('beforeunload', beforeUnload)
})
</script>
