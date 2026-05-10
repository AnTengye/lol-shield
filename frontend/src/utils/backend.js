const defaultBackendBase = 'http://127.0.0.1:9365'

export function getBackendBase({ envBase = '', isTauri = false } = {}) {
    const rawBase = envBase || (isTauri ? defaultBackendBase : '')
    return rawBase.replace(/\/$/, '')
}

export function buildBackendUrl(path, options = {}) {
    const backendBase = getBackendBase(options)
    return backendBase ? `${backendBase}${path}` : path
}

export function buildRiotAssetUrl(assetPath, options = {}) {
    return buildBackendUrl(`/riot${assetPath}`, options)
}

export function getRuntimeBackendBase() {
    return getBackendBase({
        envBase: import.meta.env.VITE_BACK_URL,
        isTauri: !!window.__TAURI_INTERNALS__,
    })
}

export function buildRuntimeBackendUrl(path) {
    return buildBackendUrl(path, {
        envBase: import.meta.env.VITE_BACK_URL,
        isTauri: !!window.__TAURI_INTERNALS__,
    })
}

export function buildRuntimeRiotAssetUrl(assetPath) {
    return buildRiotAssetUrl(assetPath, {
        envBase: import.meta.env.VITE_BACK_URL,
        isTauri: !!window.__TAURI_INTERNALS__,
    })
}
