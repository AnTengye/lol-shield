import { requestV2 } from './client'

export function getRunningSnapshotV2() {
    return requestV2.get('/running/snapshot')
}
