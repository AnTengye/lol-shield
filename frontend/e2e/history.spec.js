import { test, expect } from '@playwright/test'
import { spawn, execFileSync } from 'node:child_process'
import { mkdtemp, cp, readFile, writeFile, readdir, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

const repo = fileURLToPath(new URL('../../', import.meta.url))
const backend = 'http://127.0.0.1:9366'
let dir,
  mock,
  sidecar,
  executable,
  config,
  player,
  logs = ''
function start(file, args) {
  const child = spawn(file, args, {
    cwd: repo,
    stdio: ['ignore', 'pipe', 'pipe'],
  })
  child.stdout.on('data', (data) => {
    logs += data
  })
  child.stderr.on('data', (data) => {
    logs += data
  })
  return child
}
async function stop(child) {
  if (!child || child.exitCode !== null) return
  const exited = new Promise((resolve) => child.once('exit', resolve))
  const timer = setTimeout(() => child.kill('SIGKILL'), 5000)
  child.kill('SIGINT')
  await exited
  clearTimeout(timer)
}
async function waitStatus(predicate) {
  await expect
    .poll(
      async () => {
        try {
          return predicate(
            (await (await fetch(`${backend}/v1/status`)).json()).data,
          )
        } catch {
          return false
        }
      },
      { timeout: 30000, message: '等待独立测试 sidecar 就绪' },
    )
    .toBe(true)
}

test.beforeAll(async () => {
  dir = await mkdtemp(join(tmpdir(), 'shield-e2e-'))
  const scenario = join(dir, 'scenario')
  await cp(join(repo, 'internal/mocklcu/fixtures/default'), scenario, {
    recursive: true,
  })
  player = JSON.parse(
    await readFile(join(scenario, 'current-summoner.json'), 'utf8'),
  )
  const products = join(scenario, 'match-history/products')
  const session = JSON.parse(
    await readFile(join(scenario, 'gameflow-session.json'), 'utf8'),
  )
  const roster = [...session.gameData.teamOne]
  for (let index = roster.length; index < 10; index++) {
    const puuid = `00000000-0000-4000-8000-${String(index).padStart(12, '0')}`
    const gameName = `用于验证长名称展示的召唤师${index}`
    const summonerId = 8000000000 + index
    roster.push({
      ...roster[0],
      puuid,
      summonerId,
      summonerName: gameName,
      summonerInternalName: gameName,
      teamParticipantId: index + 1,
    })
    await cp(join(products, player.puuid), join(products, puuid), {
      recursive: true,
    })
    await writeFile(
      join(scenario, 'summoners/by-puuid', `${puuid}.json`),
      JSON.stringify({
        ...player,
        puuid,
        summonerId,
        gameName,
        tagLine: '测试',
      }),
    )
    await cp(
      join(scenario, 'ranked-stats', `${player.puuid}.json`),
      join(scenario, 'ranked-stats', `${puuid}.json`),
    )
  }
  session.gameData.teamOne = roster.slice(0, 5)
  session.gameData.teamTwo = roster.slice(5)
  session.gameData.playerChampionSelections = roster.map((member) => ({
    ...session.gameData.playerChampionSelections[0],
    championId: member.championId,
    summonerInternalName: member.summonerInternalName,
  }))
  await writeFile(
    join(scenario, 'gameflow-session.json'),
    JSON.stringify(session),
  )
  // 由真实 fixture 构造三页测试窗口，不修改仓库原始样本。
  for (const puuid of await readdir(products)) {
    const folder = join(products, puuid)
    const files = await readdir(folder)
    const sample = JSON.parse(await readFile(join(folder, files[0]), 'utf8'))
    const games = Array.from({ length: 45 }, (_, index) => ({
      ...sample.games.games[index % sample.games.games.length],
      gameId: 10913327389 + index,
      gameCreation: 1750000000000 - index * 86400000,
    }))
    for (const [begin, end] of [
      [0, 19],
      [20, 39],
      [40, 59],
    ]) {
      await writeFile(
        join(folder, `beg-${begin}-end-${end}.json`),
        JSON.stringify({
          ...sample,
          games: {
            ...sample.games,
            gameCount: 45,
            gameIndexBegin: begin,
            gameIndexEnd: end,
            games: games.slice(begin, end + 1),
          },
        }),
      )
    }
  }
  const details = join(scenario, 'match-history/games')
  const detail = JSON.parse(
    await readFile(join(details, '10913327389.json'), 'utf8'),
  )
  detail.participantIdentities = roster.map((member, index) => ({
    participantId: index + 1,
    player: {
      ...detail.participantIdentities[0].player,
      puuid: member.puuid,
      gameName: member.summonerName,
      tagLine: '测试',
    },
  }))
  const sampleParticipant = detail.participants[0]
  detail.participants = Array.from({ length: 10 }, (_, index) => ({
    ...sampleParticipant,
    participantId: index + 1,
    teamId: index < 5 ? 100 : 200,
  }))
  detail.teams = [
    { teamId: 100, win: 'Fail' },
    { teamId: 200, win: 'Win' },
  ]
  for (const index of [0, 1, 20])
    await writeFile(
      join(details, `${10913327389 + index}.json`),
      JSON.stringify({ ...detail, gameId: 10913327389 + index }),
    )
  const suffix = process.platform === 'win32' ? '.exe' : ''
  const mockExe = join(dir, `mock${suffix}`)
  executable = join(dir, `shield${suffix}`)
  execFileSync('go', ['build', '-o', mockExe, './cmd/mock-lcu'], { cwd: repo })
  execFileSync('go', ['build', '-o', executable, './cmd/shield'], { cwd: repo })
  config = join(dir, 'config.yaml')
  await writeFile(
    config,
    `web:\n  addr: 127.0.0.1:9366\nmock_lcu:\n  enabled: true\n  base_url: http://127.0.0.1:19366\n  scenario: e2e\nlog:\n  filepath: ${JSON.stringify(join(dir, 'logs'))}\n  level: error\ngame:\n  auto_confirm: false\n`,
  )
  mock = start(mockExe, [
    '--addr',
    '127.0.0.1:19366',
    '--scenario-dir',
    scenario,
  ])
  sidecar = start(executable, ['-c', config, '--data-dir', dir])
  await waitStatus((status) => status.Status === 1 && status.GameStatus === 2)
})
test.afterAll(async () => {
  await stop(sidecar)
  await stop(mock)
  if (dir) await rm(dir, { recursive: true, force: true })
})
test.afterEach(async ({}, info) => {
  if (info.status !== info.expectedStatus)
    await info.attach('sidecar-logs', { body: logs, contentType: 'text/plain' })
})

test('真实 sidecar：实时详情、分页返回、离线重启、容量管理与窗口布局', async ({
  page,
  request,
}, testInfo) => {
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('requestfailed', (req) => {
    logs += `\n请求失败 ${req.url()}: ${req.failure()?.errorText}`
  })
  page.on('console', (msg) => {
    if (msg.type() === 'error') logs += `\n浏览器: ${msg.text()}`
  })
  await page.goto('/running')
  await expect(page.locator('.history-action')).toHaveCount(10)
  await expect(
    page.locator('.team-row').first().locator('.history-action'),
  ).toHaveCount(5)
  await expect(
    page.locator('.team-row').last().locator('.history-action'),
  ).toHaveCount(5)
  await page.locator('.history-action').first().focus()
  await page.keyboard.press('Enter')
  await expect(page.locator('.history-row')).toHaveCount(20)
  expect(await page.locator('.detail-header').count()).toBe(0)
  await page.locator('.history-row').first().click()
  await expect(page.locator('.detail-header')).toBeVisible()
  await expect(page.locator('.cache-label')).toContainText('可离线查看')
  // 全部数据项默认可见，且全场最高值仅用颜色高亮，无需额外点击展开。
  await expect(page.locator('.participant-highlights').first()).toBeVisible()
  // 本场景 10 人数据完全一致，7 个高亮维度应全部命中。
  expect(await page.locator('.participant .stat-top').count()).toBe(70)
  // 视觉验证：全场最高值呈现金色，与普通数值色差显著。
  expect(
    await page
      .locator('.participant .stat-top')
      .first()
      .evaluate((el) => getComputedStyle(el).color),
  ).toBe('rgb(247, 207, 95)')
  // WeGame 风格评分：每名玩家都有评分，全场仅胜方一枚 MVP、败方一枚 SVP。
  await expect(page.locator('.participant .score-pill')).toHaveCount(10)
  await expect(page.locator('.mvp-badge')).toHaveCount(1)
  await expect(page.locator('.svp-badge')).toHaveCount(1)
  // 海克斯大乱斗（KIWI）：每位玩家展示 3 个已选强化。
  await expect(page.locator('.participant .augment')).toHaveCount(30)
  await page.locator('.ant-drawer-close').focus()
  await page.keyboard.press('Escape')
  await expect(page.locator('.detail-header')).toHaveCount(0)
  await page.locator('.history-pagination [title="2"]').click()
  await expect(
    page.locator('.history-pagination .ant-pagination-item-active'),
  ).toHaveText('2')
  await page.locator('.history-row').first().click()
  await expect(page.locator('.detail-header')).toBeVisible()
  await page.getByRole('button', { name: '返回战绩列表', exact: true }).click()
  await expect(
    page.locator('.history-pagination .ant-pagination-item-active'),
  ).toHaveText('2')
  await page.locator('.ant-drawer-close').click()
  await expect(page.locator('.history-action').first()).toBeFocused()
  await page.screenshot({
    path: testInfo.outputPath('realtime-workbench.png'),
    fullPage: true,
  })
  // 不同入口均保持正常流布局，只有内容区域自身滚动。
  for (const viewport of [
    { width: 960, height: 640 },
    { width: 1200, height: 800 },
    { width: 1440, height: 900 },
    { width: 1920, height: 1080 },
  ]) {
    await page.setViewportSize(viewport)
    for (const path of ['/running', '/rank', '/archive', '/settings']) {
      await page.goto(path)
      await expect(page.locator('h1')).toBeVisible()
      if (path === '/running') {
        await expect(page.locator('.history-action')).toHaveCount(10)
        if (viewport.width === 960) {
          await page.screenshot({
            path: testInfo.outputPath('realtime-compact.png'),
            fullPage: true,
          })
          await page.locator('.history-action').first().click()
          await page.locator('.history-row').first().focus()
          await page.keyboard.press('Space')
          await expect(page.locator('.detail-header')).toBeVisible()
          await expect(page.locator('.history-scroll')).not.toBeVisible()
          await page.keyboard.press('Escape')
          await expect(page.locator('.history-row').first()).toBeVisible()
          await page.keyboard.press('Escape')
          await expect(page.locator('.history-action').first()).toBeFocused()
        }
      }
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
      ).toBe(true)
      expect(
        await page
          .locator('.main-content')
          .evaluate((el) => el.scrollWidth <= el.clientWidth),
      ).toBe(true)
    }
  }
  const playersResponse = await request.get(`${backend}/v1/archive/players`)
  const players = (await playersResponse.json()).data.list
  expect(players.length).toBeGreaterThan(0)
  const saved = players.find((p) => p.puuid === player.puuid) || players[0]
  expect(saved.summaries).toBeGreaterThanOrEqual(40)
  await stop(mock)
  mock = null
  await waitStatus((status) => status.Status !== 1)
  await stop(sidecar)
  sidecar = start(executable, ['-c', config, '--data-dir', dir])
  await waitStatus((status) => status.Status !== 1)
  await page.setViewportSize({ width: 1200, height: 800 })
  await page.goto('/archive')
  await page.locator('.archive-controls .ant-select').first().click()
  await page
    .locator('.ant-select-item-option')
    .filter({ hasText: saved.gameName })
    .first()
    .click()
  await expect(page.locator('.history-row').first()).toBeVisible()
  await page.getByText('仅已保存详情', { exact: true }).click()
  await page.locator('.history-row').first().click()
  await expect(page.locator('.cache-label')).toContainText('本地记录')
  await page.screenshot({
    path: testInfo.outputPath('offline-history.png'),
    fullPage: true,
  })
  const stats = (await (await request.get(`${backend}/v1/cache/stats`)).json())
    .data
  expect(stats.usedBytes).toBeLessThanOrEqual(500000000)
  expect(stats.details).toBeGreaterThan(0)
  expect(
    (
      await request.post(`${backend}/v1/cache/clear`, { data: { kind: 'all' } })
    ).status(),
  ).toBe(403)
  await page.goto('/settings')
  await page.locator('.setting-row .ant-switch').first().click()
  await page.locator('nav a[href="/archive"]').click()
  await expect(
    page.getByText('尚有未保存的配置', { exact: true }),
  ).toBeVisible()
  await page.getByRole('button', { name: '继续编辑', exact: true }).click()
  await expect(page).toHaveURL(/\/settings$/)
  await page.locator('nav a[href="/archive"]').click()
  await page.getByRole('button', { name: '丢弃并离开', exact: true }).click()
  await expect(page).toHaveURL(/\/archive$/)
  await page.goto('/settings')
  await page.getByRole('button', { name: '清空战绩缓存', exact: true }).click()
  await page.getByRole('button', { name: '确认清理', exact: true }).click()
  await expect(
    page.getByText('0 位玩家 · 0 条摘要 · 0 场详情 · 0 张图片'),
  ).toBeVisible()
  expect(errors, logs).toEqual([])
})
