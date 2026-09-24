import { request } from '@/utils/request'
import qs from 'qs'

const config = ({ signal, ...params } = {}) => ({ signal, params })
export const getConfig = () => request.get('/config')
export const getVersion = () => request.get('/version')
export const getOnline = () => request.get('/status')
export const getStatus = () => request.get('/status')
export const updateConfig = (auto_confirm, auto_pick, auto_ban) => request.post('/config', { auto_confirm, auto_pick: auto_pick || 0, auto_ban: auto_ban || 0 })
export const getUser = () => request.get('/user')
export const getPlayer = (puuid, options) => request.get(`/players/${encodeURIComponent(puuid)}`, config(options))
export const getGameList = (puuid, page = 0, pageSize = 20, options = {}) => request.get(`/history/${encodeURIComponent(puuid)}`, config({ ...options, page, pageSize }))
export const getGameDetail = (gameId, options) => request.get(`/game/${gameId}`, config(options))
export const getGameRankHighest = (puuid, options) => request.get(`/rank/highest/${encodeURIComponent(puuid)}`, config(options))
export function getMulGameRankHighest(puuids, options = {}) {
  return request.get('/rank/highest', { ...config({ ...options, puuid: [puuids].flat() }), paramsSerializer: params => qs.stringify(params, { arrayFormat: 'repeat' }) })
}
export const getGameRunning = options => request.get('/game/running', config(options))
export const getSkins = () => request.get('/skins')
export const getArchivePlayers = options => request.get('/archive/players', config(options))
export const getArchiveHistory = options => request.get('/archive/history', config(options))
export const getCacheStats = () => request.get('/cache/stats')
export const clearCache = kind => request.post('/cache/clear', { kind })
