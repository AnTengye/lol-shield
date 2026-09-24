<template>
  <div v-if="visible" class="update-notice">
    <div class="update-text">
      <strong>发现新版本 v{{ update.version }}</strong>
      <span :class="failed ? 'update-error' : 'muted'">{{ hint }}</span>
    </div>
    <a-progress
      v-if="busy"
      class="update-progress"
      :percent="progress"
      :show-info="false"
      size="small"
    />
    <div class="update-actions">
      <a-button size="small" type="primary" :loading="busy" @click="install">{{
        actionLabel
      }}</a-button>
      <a-button size="small" :disabled="busy" @click="dismiss">稍后</a-button>
    </div>
  </div>
</template>
<script setup>
import { computed } from 'vue'
import { useStore } from 'vuex'
const store = useStore()
const update = computed(() => store.state.ui.update || { version: '' })
const busy = computed(() => store.getters['ui/updateBusy'])
const failed = computed(() => store.state.ui.updateStatus === 'failed')
const progress = computed(() => store.state.ui.updateProgress)
const visible = computed(
  () =>
    store.getters['ui/updateVisible'] &&
    store.state.ui.updateStatus !== 'checking',
)
const hint = computed(
  () =>
    store.state.ui.updateDetail ||
    store.state.ui.updateError ||
    '下载并安装到当前电脑，完成后自动重启。',
)
const actionLabel = computed(() =>
  failed.value ? '重试更新' : busy.value ? '更新中' : '一键更新',
)
const install = () => store.dispatch('ui/installUpdate')
const dismiss = () => store.dispatch('ui/dismissUpdate')
</script>