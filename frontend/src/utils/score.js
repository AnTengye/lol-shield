// WeGame 风格单局综合评分。
// WeGame 未公开评分公式，本实现参考其展示维度（伤害、承伤、治疗、经济、参团率、KDA，
// 其中 KDA 权重大头），对每个维度取全场百分位排名，再按权重求和得到 0-10 分。
// 百分位排名天然规避极端数值（如 40 杀的 KDA）拉分，缺失字段按 0 处理。
export const SCORE_WEIGHTS = {
    kda: 0.4,
    damage: 0.15,
    tank: 0.15,
    heal: 0.1,
    gold: 0.1,
    kp: 0.1,
}

const num = (value) => (Number.isFinite(value) ? value : 0)

export function extractScoreMetrics(player, teamKills = 0) {
    const stats = player?.stats || {}
    const kills = num(stats.kills)
    const deaths = num(stats.deaths)
    const assists = num(stats.assists)
    return {
        kda: (kills + assists) / Math.max(1, deaths),
        damage: num(stats.totalDamageDealtToChampions),
        tank: num(stats.totalDamageTaken),
        heal: num(stats.totalHeal),
        gold: num(stats.goldEarned),
        kp: teamKills > 0 ? (kills + assists) / teamKills : 0,
    }
}

// 每个维度返回 0-1 的百分位得分：并列同名次，无区分度（或单样本）取中性 0.5。
function percentileScores(values) {
    const total = values.length
    if (!total) return []
    const max = Math.max(...values)
    const min = Math.min(...values)
    if (total === 1 || max === min) return values.map(() => 0.5)
    const sorted = [...values].sort((a, b) => b - a)
    return values.map((value) => 1 - sorted.indexOf(value) / (total - 1))
}

function bestOf(group, result) {
    return group.reduce(
        (best, player) =>
            !best || result[player.participantId].score > result[best.participantId].score
                ? player
                : best,
        null,
    )
}

// players 需带 participantId / teamId / stats；teams 提供 id 与 win（true/false/null）。
// 返回 { [participantId]: { score, mvp, svp, metrics } }，胜方最高分标记 MVP，败方最高分标记 SVP。
export function computeMatchScores(players = [], teams = []) {
    const list = players.filter(Boolean)
    const teamKills = new Map()
    for (const player of list)
        teamKills.set(
            player.teamId,
            (teamKills.get(player.teamId) || 0) + num(player.stats?.kills),
        )
    const metrics = list.map((player) =>
        extractScoreMetrics(player, teamKills.get(player.teamId) || 0),
    )
    const dimensions = Object.keys(SCORE_WEIGHTS)
    const ranks = {}
    for (const dimension of dimensions)
        ranks[dimension] = percentileScores(metrics.map((item) => item[dimension]))
    const result = {}
    list.forEach((player, index) => {
        const weighted = dimensions.reduce(
            (sum, dimension) =>
                sum + SCORE_WEIGHTS[dimension] * (ranks[dimension][index] ?? 0),
            0,
        )
        result[player.participantId] = {
            score: Math.round(weighted * 100) / 10,
            mvp: false,
            svp: false,
            metrics: metrics[index],
        }
    })
    const winOf = new Map(teams.map((team) => [team.id, team.win]))
    const winners = list.filter((player) => winOf.get(player.teamId) === true)
    const losers = list.filter((player) => winOf.get(player.teamId) === false)
    const hasResult = winOf.size > 0 && [...winOf.values()].some((win) => win !== null)
    const mvp = hasResult ? bestOf(winners, result) : bestOf(list, result)
    const svp = hasResult ? bestOf(losers, result) : null
    if (mvp) result[mvp.participantId].mvp = true
    if (svp && (!mvp || svp.participantId !== mvp.participantId))
        result[svp.participantId].svp = true
    return result
}

// 评分色阶：高亮档位用于姓名旁的评分徽章。
export function scoreTier(score) {
    if (!Number.isFinite(score)) return 'none'
    if (score >= 9) return 'gold'
    if (score >= 7) return 'good'
    if (score >= 5) return 'fair'
    return 'low'
}