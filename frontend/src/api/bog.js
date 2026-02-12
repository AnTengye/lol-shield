import { request, requestMock } from '@/utils/request'
import qs from 'qs'

export function getConfig() {
    return request.get('/config', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}
export function getVersion() {
    return request.get('/version', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}
export function getOnline() {
    return request.get('/lcu', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}
export function updateConfig(auto_confirm, auto_pick, auto_ban) {
    return request({
        url: '/config',
        method: 'post',
        data: { auto_confirm, auto_pick, auto_ban }
    })
}

export function getUser() {
    return request.get('/user', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}

export function getGameList(uuid, page = 0, pageSize = 9) {
    return request.get(`/history/${uuid}`, {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        },
        params: {
            page,
            pageSize
        }
    });
}

export function getGameDetail(gameId) {
    return request.get('/game/' + gameId, {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}
export function getGameRankHighest(puuid) {
    return request.get('rank/highest/' + puuid, {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}

export function getMulGameRankHighest(...puuid) {
    return request.get('rank/highest', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        },
        params: {
            puuid,
        },
        paramsSerializer: function (params) {
            return qs.stringify(params, { arrayFormat: 'repeat' })
        }
    })
}

export function getGameRunning() {
    return request.get('game/running', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        }
    })
}


export function getSkins() {
    return request.get('skins', {
        headers: {
            'Content-Type': 'application/json;charset=UTF-8'
        },
    })
}