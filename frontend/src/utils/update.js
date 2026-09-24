export function isDesktopShell(scope) {
    return Boolean(scope && scope.__TAURI_INTERNALS__)
}

export function isWindowsAgent(agent) {
    return /windows/i.test(String(agent || ''))
}

export function formatBytes(bytes) {
    if (!Number.isFinite(bytes) || bytes <= 0) return ''
    const units = ['B', 'KB', 'MB', 'GB']
    let value = bytes
    let unit = 0
    while (value >= 1024 && unit < units.length - 1) {
        value /= 1024
        unit += 1
    }
    return `${value.toFixed(unit === 0 || value >= 100 ? 0 : 1)} ${units[unit]}`
}

export function progressPercent(received, total) {
    if (!Number.isFinite(received) || !Number.isFinite(total) || total <= 0) {
        return 0
    }
    return Math.max(0, Math.min(100, Math.round((received / total) * 100)))
}

export function describeUpdateError(error) {
    if (!error) return '更新失败，请稍后重试。'
    const message =
        typeof error === 'string'
            ? error.trim()
            : String(error.message || '').trim()
    if (!message) return '更新失败，请稍后重试。'
    if (/targets? not found/i.test(message)) {
        return '当前平台暂未提供更新包，请前往 GitHub Releases 手动下载。'
    }
    return message
}