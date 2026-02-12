const backendBase = (import.meta.env.VITE_BACK_URL || '').replace(/\/$/, '');
const wsUrl = (() => {
    if (import.meta.env.VITE_WS_URL) {
        return import.meta.env.VITE_WS_URL;
    }
    const base = backendBase || window.location.origin;
    const u = new URL(base, window.location.origin);
    const protocol = u.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${u.host}/ws`;
})();

var ws = null; // WebSocket 对象
var heartbeatTimer = null; // 心跳定时器
var isReconnect = true; // 是否自动重连
let storeRef = null;
// 创建 WebSocket 连接
export const createWebSocket = (store) => {
    storeRef = store;
    if ("WebSocket" in window) {
        ws = new WebSocket(wsUrl);
        // WebSocket 打开事件
        ws.onopen = function () {
            console.log("已连接后端");

            // 开始心跳定时器
            startHeartbeat();
        };

        // WebSocket 收到消息事件
        ws.onmessage = function (msg) {
            if (store) {
                store.commit('ws/setWsRes', JSON.parse(msg.data ?? '{}'));
            }
        };

        // 发生错误回调
        ws.onerror = function (e) {
            console.log('ws错误:', e);
        }

        // WebSocket 关闭事件
        ws.onclose = function () {
            console.log("已关闭后端连接");

            // 停止心跳定时器
            stopHeartbeat();

            if (store) {
                store.commit('ws/reset');
            }

            // 断线后自动重连
            if (isReconnect) {
                setTimeout(function () {
                    console.log("尝试重新连接后端");
                    createWebSocket(storeRef);
                }, 3 * 1000);
            }
        };
    } else {
        console.log("该浏览器不支持 WebSocket");
    }
}

// // 发送消息
// function sendMessage(message) {
//     if (ws != null && ws.readyState == WebSocket.OPEN) {
//         ws.send(message);
//         console.log("WebSocket 发送消息：" + message);
//     } else {
//         console.log("WebSocket 连接没有建立或已关闭");
//     }
// }

// 开始心跳定时器
function startHeartbeat(interval) {
    interval = interval || 30;
    heartbeatTimer = setInterval(function () {
        sendPing({ op: 1 });
    }, interval * 1000);
}

const sendPing = (message) => {
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message));
    } else {
        console.log('心跳检测失败');
    }
};

// 停止心跳定时器
function stopHeartbeat() {
    clearInterval(heartbeatTimer);
}
