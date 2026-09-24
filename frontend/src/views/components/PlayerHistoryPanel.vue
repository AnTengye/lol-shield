<template>
  <a-drawer
    :open="open"
    :width="panelWidth"
    title="召唤师战绩"
    :keyboard="false"
    :body-style="{ padding: 0, overflow: 'hidden' }"
    @close="close"
    @after-open-change="restoreFocus"
  >
    <div v-if="open" class="drawer-workspace">
      <HistoryWorkspace
        ref="workspace"
        :player="player"
        :compact="width < 1100 || !hasDetail"
        :state-key="`drawer:${player.puuid}`"
        @selection-change="hasDetail = !!$event.gameId"
      >
        <template #actions="{ current }"
          ><a-button size="small" @click="openCenter(current)"
            >在战绩中心打开</a-button
          ></template
        >
      </HistoryWorkspace>
    </div>
  </a-drawer>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useWindowWidth } from '@/composables/useWindowWidth'
import HistoryWorkspace from './HistoryWorkspace.vue'
const props = defineProps({ open: Boolean, player: { default: () => ({}) } })
const emit = defineEmits(['update:open'])
const width = useWindowWidth()
const workspace = ref(null)
const hasDetail = ref(false)
const router = useRouter()
let trigger
const panelWidth = computed(() =>
  Math.min(width.value - 32, hasDetail.value ? 1160 : 440),
)
watch(
  () => props.open,
  (open, _, onCleanup) => {
    if (!open) return
    trigger = document.activeElement
    hasDetail.value = false
    const onKey = (event) => {
      if (event.key !== 'Escape' || event.isComposing) return
      event.preventDefault()
      event.stopPropagation()
      if (!workspace.value?.back()) close()
    }
    document.addEventListener('keydown', onKey, true)
    onCleanup(() => document.removeEventListener('keydown', onKey, true))
  },
  { immediate: true },
)
const close = () => emit('update:open', false)
const restoreFocus = (open) => {
  if (!open) trigger?.focus?.()
}
function openCenter(current) {
  router.push({
    path: '/rank',
    query: {
      puuid: current.puuid,
      name: current.name,
      scope: current.scope,
      gameId: current.gameId || undefined,
      from: '/running',
    },
  })
  close()
}
</script>
