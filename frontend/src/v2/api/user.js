import { requestV2 } from './client'

export function getUserV2() {
    return requestV2.get('/user')
}

export function getSkinsV2() {
    return requestV2.get('/skins')
}
