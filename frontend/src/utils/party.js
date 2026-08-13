const PARTY_PALETTE = [
    { color: '#f0a35b', softColor: 'rgba(240, 163, 91, 0.24)' },
    { color: '#55c7d6', softColor: 'rgba(85, 199, 214, 0.24)' },
    { color: '#b38cff', softColor: 'rgba(179, 140, 255, 0.24)' },
    { color: '#e978a4', softColor: 'rgba(233, 120, 164, 0.24)' },
    { color: '#79d36e', softColor: 'rgba(121, 211, 110, 0.24)' },
    { color: '#f2d15f', softColor: 'rgba(242, 209, 95, 0.24)' },
]

const SOLO_STYLE = {
    color: '#9aa9b1',
    softColor: 'rgba(154, 169, 177, 0.18)',
}

const UNKNOWN_STYLE = {
    color: '#c3cbd0',
    softColor: 'rgba(195, 203, 208, 0.12)',
}

function compareGroupKeys(left, right) {
    const leftNumber = Number(left)
    const rightNumber = Number(right)
    const bothNumbers = Number.isFinite(leftNumber) && Number.isFinite(rightNumber)

    if (bothNumbers && leftNumber !== rightNumber) {
        return leftNumber - rightNumber
    }

    return String(left).localeCompare(String(right))
}

function groupLabel(index) {
    let value = index
    let label = ''

    do {
        label = String.fromCharCode(65 + (value % 26)) + label
        value = Math.floor(value / 26) - 1
    } while (value >= 0)

    return `组队 ${label}`
}

function normalizeMembers(members) {
    if (!Array.isArray(members)) {
        return []
    }

    const seen = new Set()
    return members.filter((member) => {
        const puuid = member?.puuid
        if (!puuid || seen.has(puuid)) {
            return false
        }

        seen.add(puuid)
        return true
    })
}

function createPartyMemberMeta({ key, label, size, order, style, kind }) {
    return {
        partyKey: key,
        partyLabel: label,
        partySize: size,
        partyOrder: order,
        partyKind: kind,
        partyColor: style.color,
        partySoftColor: style.softColor,
    }
}

/**
 * Converts the backend's preTeam map into explicit card metadata.
 * A one-player backend group is deliberately rendered as "单排" instead of
 * sharing a color with another one-player group.
 */
export function buildPartyView(preTeam = {}, players = []) {
    const groups = Object.entries(preTeam ?? {})
        .map(([key, members]) => ({ key, members: normalizeMembers(members) }))
        .filter((group) => group.members.length > 0)
        .sort((left, right) => compareGroupKeys(left.key, right.key))

    const byPuuid = {}
    const summaries = []
    let soloCount = 0
    let partyOrder = 0

    for (const group of groups) {
        if (group.members.length > 1) {
            const style = PARTY_PALETTE[partyOrder % PARTY_PALETTE.length]
            const key = `party:${group.key}`
            const label = groupLabel(partyOrder)
            const meta = createPartyMemberMeta({
                key,
                label,
                size: group.members.length,
                order: partyOrder,
                style,
                kind: 'party',
            })

            summaries.push({ ...meta, size: group.members.length })
            for (const member of group.members) {
                byPuuid[member.puuid] = { ...meta }
            }
            partyOrder++
            continue
        }

        const member = group.members[0]
        byPuuid[member.puuid] = createPartyMemberMeta({
            key: `solo:${member.puuid}`,
            label: '单排',
            size: 1,
            order: Number.MAX_SAFE_INTEGER,
            style: SOLO_STYLE,
            kind: 'solo',
        })
        soloCount++
    }

    if (soloCount > 0) {
        summaries.push({
            ...createPartyMemberMeta({
                key: 'solo',
                label: '单排',
                size: soloCount,
                order: Number.MAX_SAFE_INTEGER,
                style: SOLO_STYLE,
                kind: 'solo',
            }),
            size: soloCount,
        })
    }

    const hasEvidence = groups.length > 0
    for (const player of players) {
        if (player?.puuid && !byPuuid[player.puuid]) {
            byPuuid[player.puuid] = createPartyMemberMeta({
                key: `unknown:${player.puuid}`,
                label: '组队信息待确认',
                size: 0,
                order: Number.MAX_SAFE_INTEGER,
                style: UNKNOWN_STYLE,
                kind: 'unknown',
            })
        }
    }

    if (!hasEvidence && players.length > 0) {
        summaries.push({
            ...createPartyMemberMeta({
                key: 'unknown',
                label: '组队信息待确认',
                size: players.length,
                order: Number.MAX_SAFE_INTEGER,
                style: UNKNOWN_STYLE,
                kind: 'unknown',
            }),
            size: players.length,
        })
    }

    return { byPuuid, summaries, hasEvidence }
}

export function summarizeParties(players = []) {
    const summaries = new Map()

    for (const player of players) {
        if (!player) {
            continue
        }

        const key = player.partyKind === 'solo'
            ? 'solo'
            : player.partyKind === 'unknown'
                ? 'unknown'
                : player.partyKey || `unknown:${player.puuid}`
        const current = summaries.get(key)

        if (current) {
            current.size++
            continue
        }

        summaries.set(key, {
            partyKey: key,
            partyLabel: player.partyLabel || '组队信息待确认',
            partySize: player.partySize || 0,
            partyOrder: player.partyOrder ?? Number.MAX_SAFE_INTEGER,
            partyKind: player.partyKind || 'unknown',
            partyColor: player.partyColor || UNKNOWN_STYLE.color,
            partySoftColor: player.partySoftColor || UNKNOWN_STYLE.softColor,
            size: 1,
        })
    }

    return [...summaries.values()].sort((left, right) => left.partyOrder - right.partyOrder)
}
