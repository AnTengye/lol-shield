import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { nextTick } from 'vue'
import UpdateNotice from './UpdateNotice.vue'
import ui from '@/store/ui'

const stubs = {
  AButton: { template: '<button><slot /></button>' },
  AProgress: {
    props: ['percent'],
    template: '<div class="progress">{{ percent }}</div>',
  },
}
const wrappers = []
let store
function render() {
  const wrapper = mount(UpdateNotice, {
    global: { plugins: [store], stubs },
  })
  wrappers.push(wrapper)
  return wrapper
}

beforeEach(() => {
  const cache = new Map()
  vi.stubGlobal('localStorage', {
    getItem: (key) => cache.get(key) ?? null,
    setItem: (key, value) => cache.set(String(key), String(value)),
    removeItem: (key) => cache.delete(key),
    clear: () => cache.clear(),
  })
  store = createStore({ modules: { ui } })
})

afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
})

describe('更新提示条', () => {
  it('没有新版本时不打断界面', () => {
    expect(render().find('.update-notice').exists()).toBe(false)
  })

  it('发现新版本时展示版本与一键更新', () => {
    store.commit('ui/updateState', { update: { version: '2.0.0-rc.12' } })
    const wrapper = render()
    expect(wrapper.text()).toContain('发现新版本 v2.0.0-rc.12')
    expect(wrapper.text()).toContain('一键更新')
  })

  it('查看对局详情期间不展示提示', () => {
    store.commit('ui/updateState', { update: { version: '2.0.0-rc.12' } })
    store.commit('ui/detailOpen', true)
    expect(render().find('.update-notice').exists()).toBe(false)
  })

  it('下载期间展示进度并锁定操作', () => {
    store.commit('ui/updateState', {
      update: { version: '2.0.0-rc.12' },
      updateStatus: 'downloading',
      updateProgress: 42,
      updateDetail: '正在下载更新（2.0 MB）',
    })
    const wrapper = render()
    expect(wrapper.text()).toContain('正在下载更新（2.0 MB）')
    expect(wrapper.find('.progress').text()).toBe('42')
    expect(wrapper.text()).toContain('更新中')
  })

  it('安装失败时提示原因并可重试', () => {
    store.commit('ui/updateState', {
      update: { version: '2.0.0-rc.12' },
      updateStatus: 'failed',
      updateError: '下载中断',
    })
    const wrapper = render()
    expect(wrapper.text()).toContain('下载中断')
    expect(wrapper.text()).toContain('重试更新')
  })

  it('一键更新触发安装流程', async () => {
    store.commit('ui/updateState', { update: { version: '2.0.0-rc.12' } })
    const dispatch = vi.spyOn(store, 'dispatch')
    const wrapper = render()
    await wrapper.findAll('button')[0].trigger('click')
    expect(dispatch).toHaveBeenCalledWith('ui/installUpdate')
  })

  it('稍后关闭提示并记住版本', async () => {
    store.commit('ui/updateState', { update: { version: '2.0.0-rc.12' } })
    const wrapper = render()
    await wrapper.findAll('button')[1].trigger('click')
    await nextTick()
    expect(store.state.ui.dismissedUpdate).toBe('2.0.0-rc.12')
    expect(wrapper.find('.update-notice').exists()).toBe(false)
  })
})