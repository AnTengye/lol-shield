import { onScopeDispose, ref, watch } from 'vue'
import { getArchiveHistory, getGameList } from '@/api/bog'

export function usePlayerHistory(selection) {
  const list = ref([])
  const total = ref(0)
  const meta = ref(null)
  const loading = ref(false)
  const error = ref('')
  let generation = 0
  let controller
  async function load(policy) {
    const current = ++generation
    controller?.abort()
    controller = new AbortController()
    const query = selection.value
    if (policy !== 'refresh') {
      list.value = []
      total.value = 0
      meta.value = null
    }
    error.value = ''
    loading.value = false
    if (!query.puuid) return
    loading.value = true
    try {
      const options = {
        ...query,
        page: query.page - 1,
        pageSize: 20,
        policy: policy || 'cache-first',
        signal: controller.signal,
      }
      const response = query.offline
        ? await getArchiveHistory(options)
        : await getGameList(query.puuid, query.page - 1, 20, options)
      if (current !== generation) return
      list.value = response.data?.list || []
      total.value = response.data?.total || 0
      meta.value = response.meta
    } catch (err) {
      if (current === generation && err.code !== 'ERR_CANCELED') {
        if (policy === 'refresh' && list.value.length)
          meta.value = { ...meta.value, refreshFailed: true }
        else error.value = err.message || '战绩加载失败'
      }
    } finally {
      if (current === generation) loading.value = false
    }
  }
  watch(selection, () => load(), { immediate: true, deep: true })
  onScopeDispose(() => {
    generation++
    controller?.abort()
  })
  return { list, total, meta, loading, error, refresh: () => load('refresh') }
}
