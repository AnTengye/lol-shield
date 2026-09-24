import { listen } from '@tauri-apps/api/event'
import { getRuntimeBackendBase } from '@/utils/backend'

const base = getRuntimeBackendBase() || window.location.origin
const url = new URL(base, window.location.origin)
const wsUrl =
  import.meta.env.VITE_WS_URL ||
  `${url.protocol === 'https:' ? 'wss:' : 'ws:'}//${url.host}/ws`
let generation = 0
let ws, heartbeat, retry
let listeners = []

export function destroyWebSocket() {
  generation++
  clearInterval(heartbeat)
  clearTimeout(retry)
  listeners.forEach((unlisten) => unlisten())
  listeners = []
  if (ws) {
    ws.onclose = ws.onmessage = ws.onopen = ws.onerror = null
    ws.close()
    ws = null
  }
}

export async function createWebSocket(store) {
  destroyWebSocket()
  const current = generation
  const active = () => current === generation
  const transport = (online) => {
    if (!active()) return
    store.commit('ui/backendOnline', online)
    if (!online) store.commit('ws/reset')
  }
  if (window.__TAURI_INTERNALS__) {
    for (const [event, callback] of [
      [
        'shield-status',
        (event) => {
          if (active()) store.commit('ws/setWsRes', event.payload ?? {})
        },
      ],
      ['shield-transport', (event) => transport(event.payload === true)],
    ]) {
      try {
        const unlisten = await listen(event, callback)
        if (active()) listeners.push(unlisten)
        else unlisten()
      } catch {
        transport(false)
      }
    }
    return
  }
  const connect = () => {
    if (!active() || !('WebSocket' in window)) return
    const socket = new WebSocket(wsUrl)
    ws = socket
    socket.onopen = () => {
      if (!active()) return
      transport(true)
      clearInterval(heartbeat)
      heartbeat = setInterval(() => {
        if (socket.readyState === WebSocket.OPEN)
          socket.send(JSON.stringify({ op: 1 }))
      }, 30000)
    }
    socket.onmessage = (event) => {
      if (!active()) return
      try {
        store.commit('ws/setWsRes', JSON.parse(event.data))
      } catch {
        /* 忽略无效状态帧 */
      }
    }
    socket.onerror = () => {
      /* 重连由关闭事件统一调度 */
    }
    socket.onclose = () => {
      if (!active()) return
      clearInterval(heartbeat)
      transport(false)
      retry = setTimeout(connect, 3000)
    }
  }
  connect()
}
