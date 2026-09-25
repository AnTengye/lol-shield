import gameItem from './items.generated.js'
import augment from './augments.generated.js'

// gameItem 字典由 frontend/scripts/generate-item-dict.mjs 生成，
// augment 字典由 frontend/scripts/generate-augment-dict.mjs 生成，
// 游戏版本更新新增或改名装备/强化后运行对应 generate 命令重新生成并提交。
export default {
    dictAttr: {
        rank: [
            { key: 'IRON', value: '坚韧黑铁' },
            { key: 'BRONZE', value: '英勇黄铜' },
            { key: 'SILVER', value: '不屈白银' },
            { key: 'GOLD', value: '荣耀黄金' },
            { key: 'PLATINUM', value: '华贵铂金' },
            { key: 'EMERALD', value: '流光翡翠' },
            { key: 'DIAMOND', value: '璀璨钻石' },
            { key: 'MASTER', value: '超凡大师' },
            { key: 'GRANDMASTER', value: '傲世宗师' },
            { key: 'CHALLENGER', value: '最强王者' }
        ],
        mode: [
            { key: 'ARAM', value: '大乱斗' },
            { key: 'CLASSIC', value: '匹配' },
            { key: 'URF', value: '无限火力' },
            { key: 'PRACTICETOOL', value: '自定义' },
            { key: 'STRAWBERRY', value: '无尽狂潮' },
        ],
        queueType: [
            { key: 'RANKED_SOLO_5x5', value: '单双排' },
            { key: 'RANKED_FLEX_SR', value: '灵活排位' },
            { key: 'NORMAL', value: '匹配' },
            { key: 'ARAM_UNRANKED_5x5', value: '大乱斗' },
            { key: 'RANKED_TFT', value: '云顶之弈' },
        ],
        queue: [
            { key: 0, value: '自定义' },
            { key: 430, value: '自选匹配' },
            { key: 420, value: '单双排' },
            { key: 440, value: '灵活组排' },
            { key: 450, value: '极地大乱斗' },
            { key: 900, value: '无限火力' },
            { key: 830, value: '人机入门' },
            { key: 840, value: '人机新手' },
            { key: 850, value: '人机一般' },
            { key: 1700, value: '斗魂竞技场' },
            { key: 1810, value: '无尽狂潮单人' },
            { key: 1820, value: '无尽狂潮双人' },
            { key: 1830, value: '无尽狂潮三人' },
            { key: 1840, value: '无尽狂潮四人' },
            { key: 1900, value: '无限乱斗' },

        ],
    },
    dict: {
        spell: {
            '1': 'Summoner_boost.png',
            '3': 'Summoner_exhaust.png',
            '4': 'Summoner_flash.png',
            '5': 'Summoner_Backtrack.png',
            '6': 'Summoner_haste.png',
            '7': 'Summoner_heal.png',
            '11': 'Summoner_smite.png',
            '12': 'Summoner_teleport.png',
            '13': 'SummonerMana.png',
            '14': 'SummonerIgnite.png',
            '21': 'SummonerBarrier.png',
            '30': 'Benevolence_Of_King_Poro_Icon.png',
            '31': 'Trailblazer_Poro_Icon.png',
            '32': 'Summoner_Mark.png',
            '39': 'Summoner_Mark.png',
            '54': 'Summoner_Empty.png',
            '55': 'Summoner_EmptySmite.png',
            '2201': 'Icon_SummonerSpell_Flee.2v2_Mode_Fighters.png',
            '2202': 'Summoner_flash.png',
            '2203': 'Summoner_flash.png',
            '4294967295': '1102_Smite.png'
        },
        gameItem,
        augment,
    },
}