import { requestV2 } from './client'

export function getSystemStateV2() {
    return requestV2.get('/system/state')
}
