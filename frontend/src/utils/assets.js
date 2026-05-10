import { buildRuntimeRiotAssetUrl } from './backend.js'

export function buildGameItemIconPath(fileName) {
    return `/ASSETS/Items/Icons2D/${fileName}`
}

export function buildRuntimeGameItemIconUrl(fileName) {
    return buildRuntimeRiotAssetUrl(buildGameItemIconPath(fileName))
}
