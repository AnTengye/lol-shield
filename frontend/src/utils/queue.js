export function resolveQueueName(queueMap, queueId, fallbackName = '') {
    const key = `${queueId}`
    if (queueMap[key]) {
        return queueMap[key]
    }
    if (fallbackName && fallbackName.trim() !== '') {
        return fallbackName
    }
    return `队列 ${queueId}`
}

export function isAramLikeQueue(queueId, queueName = '') {
    if (`${queueId}` === '450') {
        return true
    }
    return queueName.includes('大乱斗')
}
