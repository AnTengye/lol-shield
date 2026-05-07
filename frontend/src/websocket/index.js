import { listen } from '@tauri-apps/api/event'

const defaultBackendBase = 'http://127.0.0.1:9365'
const backendBase = (
    import.meta.env.VITE_BACK_URL ||
    (window.__TAURI_INTERNALS__ ? defaultBackendBase : '')
).replace(/\/$/, '')

const wsUrl = (() => {
    if (import.meta.env.VITE_WS_URL) {
        return import.meta.env.VITE_WS_URL
    }
    const base = backendBase || window.location.origin
    const u = new URL(base, window.location.origin)
    const protocol = u.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${u.host}/ws`
})()

let unlistenStatus = null
let ws = null
let heartbeatTimer = null
let isReconnect = true
let storeRef = null

export const createWebSocket = async (store) => {
    storeRef = store

    if (window.__TAURI_INTERNALS__) {
        if (unlistenStatus) {
            unlistenStatus()
        }

        unlistenStatus = await listen('shield-status', (event) => {
            if (store) {
                store.commit('ws/setWsRes', event.payload ?? {})
            }
        })
        return
    }

    createBrowserWebSocket(store)
}

const createBrowserWebSocket = (store) => {
    if (!('WebSocket' in window)) {
        console.log('该浏览器不支持 WebSocket')
        return
    }

    ws = new WebSocket(wsUrl)

    ws.onopen = function () {
        console.log('已连接后端')
        startHeartbeat()
    }

    ws.onmessage = function (msg) {
        if (store) {
            store.commit('ws/setWsRes', JSON.parse(msg.data ?? '{}'))
        }
    }

    ws.onerror = function (e) {
        console.log('ws错误:', e)
    }

    ws.onclose = function () {
        console.log('已关闭后端连接')
        stopHeartbeat()

        if (store) {
            store.commit('ws/reset')
        }

        if (isReconnect) {
            setTimeout(function () {
                console.log('尝试重新连接后端')
                createBrowserWebSocket(storeRef)
            }, 3 * 1000)
        }
    }
}

function startHeartbeat(interval) {
    interval = interval || 30
    heartbeatTimer = setInterval(function () {
        sendPing({ op: 1 })
    }, interval * 1000)
}

const sendPing = (message) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message))
    }
}

function stopHeartbeat() {
    clearInterval(heartbeatTimer)
}
