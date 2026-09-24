import { computed, onScopeDispose, ref, watch } from 'vue'
import { getGameDetail, getMulGameRankHighest } from '@/api/bog'

// 所有入口共用请求代次，旧请求（包括 finally）不能覆盖新选择。
export function useMatchDetail(selection) {
  const data = ref(null)
  const meta = ref(null)
  const ranks = ref({})
  const loading = ref(false)
  const error = ref('')
  const errorCode = ref('')
  let generation = 0
  let controller

  async function load(policy) {
    const current = ++generation
    controller?.abort()
    controller = new AbortController()
    const { gameId, scope, offline } = selection.value
    const retain = policy === 'refresh' && data.value
    if (!retain) {
      data.value = null
      ranks.value = {}
      meta.value = null
    }
    error.value = ''
    errorCode.value = ''
    loading.value = false
    if (!gameId) return
    loading.value = true
    const options = {
      scope,
      policy: offline ? 'cache-only' : policy || 'cache-first',
      signal: controller.signal,
    }
    try {
      const response = await getGameDetail(gameId, options)
      if (current !== generation) return
      data.value = response.data
      meta.value = response.meta
      const puuids = [
        ...new Set(
          (response.data.participantIdentities || [])
            .map((p) => p.player?.puuid)
            .filter(Boolean),
        ),
      ]
      if (puuids.length) {
        getMulGameRankHighest(puuids, {
          ...options,
          scope: response.meta?.scope || scope,
        })
          .then((result) => {
            if (current === generation)
              ranks.value = Object.fromEntries(
                (result.data || []).map((item) => [item.puuid, item.data]),
              )
          })
          .catch(() => {})
      }
    } catch (err) {
      if (current === generation && err.code !== 'ERR_CANCELED') {
        if (retain) meta.value = { ...meta.value, refreshFailed: true }
        else {
          errorCode.value = err.businessCode || ''
          error.value = err.message || '详情加载失败，请重试'
        }
      }
    } finally {
      if (current === generation) loading.value = false
    }
  }
  watch(selection, () => load(), { immediate: true })
  onScopeDispose(() => {
    generation++
    controller?.abort()
  })
  return {
    data,
    meta,
    ranks,
    loading,
    error,
    errorCode,
    refresh: () => load('refresh'),
    empty: computed(() => !data.value && !loading.value),
  }
}
