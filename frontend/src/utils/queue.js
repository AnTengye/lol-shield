import generatedQueues from '../model/dicts/queues.generated.js'

const RAW_GAME_MODE_PATTERN = /^[A-Z0-9_]+$/

function cleanName(name) {
    return `${name || ''}`.trim()
}

function isRawGameModeName(name) {
    return RAW_GAME_MODE_PATTERN.test(cleanName(name))
}

export function resolveQueueName(queueMap, queueId, fallbackName = '') {
    const key = `${queueId}`
    if (queueMap[key]) {
        return queueMap[key]
    }
    const fallback = cleanName(fallbackName)
    if (fallback && !isRawGameModeName(fallback)) {
        return fallback
    }
    if (generatedQueues[key]?.name) {
        return generatedQueues[key].name
    }
    if (fallback) {
        return fallback
    }
    return `队列 ${queueId}`
}

export function isAramLikeQueue(queueId, queueName = '') {
    const key = `${queueId}`
    if (generatedQueues[key]?.aramLike) {
        return true
    }
    return queueName.includes('大乱斗') || /ARAM|Howling Abyss/i.test(queueName)
}
