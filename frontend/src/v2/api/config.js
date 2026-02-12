import { requestV2 } from './client'

export function getConfigV2() {
    return requestV2.get('/config')
}

export function patchConfigV2(payload) {
    return requestV2.patch('/config', payload)
}
