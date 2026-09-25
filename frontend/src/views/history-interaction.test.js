import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { createMemoryHistory, createRouter } from 'vue-router'
import { nextTick } from 'vue'
import RealTime from './RealTime.vue'
import RankDetail from './RankDetail.vue'
import HistoryWorkspace from './components/HistoryWorkspace.vue'
import ui from '@/store/ui'
import ws from '@/store/websocket'
import * as api from '@/api/bog'

vi.mock('@/api/bog', () => ({
  getPlayer: vi.fn(),
  getGameRankHighest: vi.fn(),
  getGameList: vi.fn(),
  getArchiveHistory: vi.fn(),
  getGameDetail: vi.fn(),
  getMulGameRankHighest: vi.fn(),
  getGameRunning: vi.fn(),
}))
const scope = 'mock:default:KR'
const summary = (id) => ({
  gameId: id,
  scope,
  queueId: 450,
  championId: 1,
  createTime: 10000,
  gameDuration: 1200,
  kills: 7,
  deaths: 2,
  assists: 9,
  win: true,
})
const detail = (id) => ({
  code: '0',
  meta: { scope, source: 'lcu', cached: true },
  data: {
    gameId: id,
    queueId: 450,
    gameDuration: 1200,
    gameVersion: `版本-${id}`,
    teams: [
      { teamId: 100, win: 'Win' },
      { teamId: 300, win: 'Fail' },
    ],
    participantIdentities: [
      { participantId: 1, player: { puuid: 'p', gameName: '玩家甲' } },
      { participantId: 2, player: { puuid: 'q', gameName: '玩家乙' } },
    ],
    participants: [
      {
        participantId: 1,
        teamId: 100,
        championId: 1,
        stats: { kills: 7, deaths: 2, assists: 9 },
      },
      { participantId: 2, teamId: 300, championId: 2, stats: {} },
    ],
  },
})
const stubs = {
  AButton: { template: '<button><slot /></button>' },
  ADrawer: { props: ['open'], template: '<aside v-if="open"><slot /></aside>' },
  ASkeleton: { template: '<div>加载中</div>' },
  AAlert: { props: ['message'], template: '<div>{{ message }}</div>' },
  AEmpty: { template: '<div>空列表</div>' },
  ASwitch: true,
  APagination: true,
  APopover: { template: '<div><slot /><slot name="content" /></div>' },
  ACheckbox: true,
}
let store, router
const wrappers = []
function render(component, props = {}) {
  const wrapper = mount(component, {
    props,
    global: { plugins: [store, router], stubs },
  })
  wrappers.push(wrapper)
  return wrapper
}
const deferred = () => {
  let resolve, reject
  const promise = new Promise((a, b) => {
    resolve = a
    reject = b
  })
  return { promise, resolve, reject }
}
beforeEach(async () => {
  vi.resetAllMocks()
  const storage = new Map()
  vi.stubGlobal('localStorage', {
    getItem: (key) => storage.get(key) ?? null,
    setItem: (key, value) => storage.set(key, value),
    clear: () => storage.clear(),
  })
  store = createStore({
    modules: {
      ui,
      ws: {
        ...ws,
        state: { ...ws.state, status: 1, uuid: 'p', gameStatus: 2 },
      },
    },
  })
  router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
  await router.push('/running')
  api.getPlayer.mockResolvedValue({
    data: { gameName: '玩家甲', summonerLevel: 99 },
  })
  api.getGameRankHighest.mockResolvedValue({
    data: { tier: 'GOLD', division: 'I' },
  })
  api.getGameList.mockResolvedValue({
    data: { list: [summary(42), summary(43)], total: 40 },
    meta: { scope, cached: true },
  })
  api.getArchiveHistory.mockResolvedValue({
    data: { list: [summary(42), summary(43)], total: 40 },
    meta: { scope, source: 'disk-cache', cached: true },
  })
  api.getGameDetail.mockImplementation((id) => Promise.resolve(detail(id)))
  api.getMulGameRankHighest.mockResolvedValue({ data: [] })
  api.getGameRunning.mockResolvedValue({
    data: {
      selfTeamInfo: { teamId: 100, userList: [{ puuid: 'p' }] },
      enemyTeamInfo: { teamId: 200, userList: [{ puuid: 'q' }] },
      userNameMap: { p: { gameName: '玩家甲' }, q: { gameName: '玩家乙' } },
      allGameHistory: {},
      preTeam: {},
    },
  })
})
afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
})

describe('共用战绩交互', () => {
  it('实时玩家 → 列表 → 详情 → 参与者 → 原对局，不堆叠面板', async () => {
    const wrapper = render(RealTime)
    await flushPromises()
    expect(wrapper.text()).toContain('近 0 场胜率')
    await wrapper.find('.history-action').trigger('click')
    await flushPromises()
    expect(api.getGameDetail).not.toHaveBeenCalled()
    await wrapper.find('.history-row').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('版本-42')
    expect(wrapper.text()).toContain('队伍 300')
    await wrapper.findAll('.player-link')[1].trigger('click')
    await flushPromises()
    expect(wrapper.find('.workspace-header').text()).toContain('玩家乙')
    await wrapper.find('.workspace-header .text-button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('版本-42')
    expect(wrapper.findAll('aside')).toHaveLength(1)
    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }),
    )
    await nextTick()
    await flushPromises()
    expect(wrapper.find('.history-row').attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('.detail-header').exists()).toBe(false)
    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }),
    )
    await nextTick()
    expect(wrapper.find('aside').exists()).toBe(false)
  })
  it('返回列表、重新挂载和筛选保留或正确重置分页及滚动', async () => {
    const props = { player: { puuid: 'p' }, compact: true, stateKey: 'test' }
    const wrapper = render(HistoryWorkspace, props)
    await flushPromises()
    const list = wrapper.findComponent({ name: 'GameHistoryList' })
    list.vm.$emit('page-change', 2)
    await flushPromises()
    list.vm.$emit('scroll-change', 180)
    await wrapper.findAll('.history-row')[1].trigger('click')
    await flushPromises()
    wrapper.vm.back()
    await nextTick()
    expect(wrapper.vm.current.page).toBe(2)
    expect(wrapper.vm.current.scrollTop).toBe(180)
    wrapper.unmount()
    const restored = render(HistoryWorkspace, props)
    await flushPromises()
    expect(restored.vm.current.page).toBe(2)
    expect(restored.vm.current.scope).toBe(scope)
    expect(restored.vm.current.lastGameId).toBe(43)
    await restored.setProps({ filters: { win: 'true' } })
    expect(restored.vm.current.page).toBe(1)
    expect(restored.vm.current.scrollTop).toBe(0)
  })
  it('已识别范围的初次挂载读取资料，刷新列表失败保留内容', async () => {
    const wrapper = render(HistoryWorkspace, {
      player: { puuid: 'p', scope },
      offline: true,
    })
    await flushPromises()
    expect(api.getPlayer.mock.calls[0][1].policy).toBe('cache-only')
    expect(wrapper.text()).toContain('等级 99')
    api.getArchiveHistory.mockRejectedValue(new Error('本地服务断开'))
    await wrapper.find('.section-toolbar button').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.history-row')).toHaveLength(2)
    expect(wrapper.text()).toContain('刷新失败，保留上次查看的列表')
  })
  it('A/B 快速切换隔离旧成功、失败、finally 和段位', async () => {
    const a = deferred(),
      b = deferred(),
      rank = deferred()
    api.getGameDetail.mockImplementation((id) =>
      id === 42 ? a.promise : b.promise,
    )
    api.getMulGameRankHighest.mockReturnValue(rank.promise)
    const wrapper = render(RankDetail, { gameId: 42, scope })
    await wrapper.setProps({ gameId: 43 })
    a.resolve(detail(42))
    await flushPromises()
    expect(wrapper.text()).toContain('加载中')
    b.resolve(detail(43))
    await flushPromises()
    expect(wrapper.text()).toContain('版本-43')
    expect(wrapper.text()).not.toContain('版本-42')
    api.getGameDetail.mockResolvedValue(detail(44))
    api.getMulGameRankHighest.mockResolvedValue({ data: [] })
    await wrapper.setProps({ gameId: 44 })
    await flushPromises()
    rank.resolve({ data: [{ puuid: 'p', data: { tier: 'CHALLENGER' } }] })
    await flushPromises()
    expect(wrapper.text()).not.toContain('最近段位：最强王者')
  })
  it('段位失败不阻断详情，刷新断网保留原数据', async () => {
    api.getMulGameRankHighest.mockRejectedValue(new Error('段位不可用'))
    const wrapper = render(RankDetail, { gameId: 42, scope })
    await flushPromises()
    expect(wrapper.text()).toContain('版本-42')
    api.getGameDetail.mockRejectedValue(new Error('网络断开'))
    await wrapper.find('.detail-header button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('版本-42')
    expect(wrapper.text()).toContain('刷新失败')
  })
  it('离线摘要未保存详情有明确提示，强制 cache-only', async () => {
    api.getGameDetail.mockRejectedValue(
      Object.assign(new Error('尚未保存'), { businessCode: 'CACHE_MISS' }),
    )
    const wrapper = render(HistoryWorkspace, {
      player: { puuid: 'p', scope },
      offline: true,
    })
    await flushPromises()
    await wrapper.find('.history-row').trigger('click')
    await flushPromises()
    expect(api.getGameDetail.mock.calls[0][1].policy).toBe('cache-only')
    expect(wrapper.text()).toContain('本地只有摘要')
    expect(api.getGameList).not.toHaveBeenCalled()
  })
  it('缺失统计不变成零，未知结果不当作失败', async () => {
    const response = detail(42)
    response.data.teams[1].win = ''
    api.getGameDetail.mockResolvedValue(response)
    const wrapper = render(RankDetail, { gameId: 42, scope })
    await flushPromises()
    expect(wrapper.findAll('.team-summary')[1].text()).toContain('结果未知')
    expect(wrapper.findAll('.team-summary')[1].text()).toContain('— 击杀')
    expect(wrapper.findAll('.team-summary')[1].text()).toContain('— 金币')
    expect(wrapper.findAll('.participant-kda')[1].text()).toContain('—/—/—')
  })
  it('详情默认展示全部数据，并高亮全场最高值', async () => {
    const rich = detail(42)
    rich.data.participants = [
      {
        participantId: 1,
        teamId: 100,
        championId: 1,
        stats: {
          kills: 7,
          deaths: 2,
          assists: 9,
          goldEarned: 9000,
          totalDamageDealtToChampions: 12000,
          totalDamageTaken: 21000,
          damageDealtToObjectives: 3000,
          timeCCingOthers: 12.4,
          visionScore: 20,
          pentaKills: 1,
          playerAugment1: 1009,
          playerAugment2: 1195,
          playerAugment3: 1320,
        },
      },
      {
        participantId: 2,
        teamId: 300,
        championId: 2,
        stats: {
          kills: 3,
          deaths: 5,
          assists: 4,
          goldEarned: 15000,
          totalDamageDealtToChampions: 30000,
          totalDamageTaken: 9000,
          damageDealtToObjectives: 8000,
          timeCCingOthers: 30.2,
          visionScore: 45,
          pentaKills: 0,
        },
      },
    ]
    api.getGameDetail.mockResolvedValue(rich)
    const wrapper = render(RankDetail, { gameId: 42, scope })
    await flushPromises()
    const rows = wrapper.findAll('.participant')
    expect(rows).toHaveLength(2)
    // 无需任何点击即可看到全部数据项
    for (const label of ['伤害', '经济', '承伤', '目标伤害', '控制', '视野'])
      expect(rows[0].text()).toContain(label)
    expect(rows[0].text()).toContain('控制 12秒')
    // 全场最高只标在对应玩家上，仅用颜色区分，不再输出“最高 xx”文字
    expect(rows[0].findAll('.stat-top')).toHaveLength(2)
    expect(rows[1].findAll('.stat-top')).toHaveLength(5)
    expect(rows[0].text()).not.toContain('最高')
    expect(rows[0].text()).toContain('五杀')
    expect(rows[1].text()).not.toContain('五杀')
    // WeGame 风格评分：胜方最高分 MVP、败方最高分 SVP
    expect(rows[0].find('.score-pill').text()).toBe('6.0')
    expect(rows[1].find('.score-pill').text()).toBe('4.0')
    expect(rows[0].find('.mvp-badge').exists()).toBe(true)
    expect(rows[1].find('.svp-badge').exists()).toBe(true)
    // 海克斯大乱斗强化：直接展示玩家选取的强化图标
    expect(rows[0].findAll('.augment')).toHaveLength(3)
    expect(rows[0].findAll('.augment')[0].attributes('title')).toBe('霸符兄弟')
    expect(rows[1].findAll('.augment')).toHaveLength(0)
  })
  it('离线服务故障保留真实错误，不把有效详情标为淘汰', async () => {
    api.getGameDetail.mockRejectedValue(new Error('本地服务连接失败'))
    const wrapper = render(RankDetail, { gameId: 42, scope, offline: true })
    await flushPromises()
    expect(wrapper.text()).toContain('本地服务连接失败')
    expect(wrapper.text()).not.toContain('本地只有摘要')
    expect(wrapper.emitted('availability-change')).toBeUndefined()
  })
  it('清空后取消旧视图且不自动重新抓取', async () => {
    const wrapper = render(HistoryWorkspace, { player: { puuid: 'p', scope } })
    await flushPromises()
    await wrapper.find('.history-row').trigger('click')
    await flushPromises()
    const calls = api.getGameList.mock.calls.length
    store.commit('ui/cacheCleared')
    await flushPromises()
    expect(api.getGameList.mock.calls.length).toBe(calls)
    expect(wrapper.findAll('.history-row')).toHaveLength(0)
    expect(wrapper.text()).not.toContain('版本-42')
  })
})
