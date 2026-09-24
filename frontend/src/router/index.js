import { createRouter, createWebHashHistory, createWebHistory } from 'vue-router'
import SettingPage from '../views/SettingFrom.vue'
import RankList from '../views/RankList.vue'
import RealTime from '../views/RealTime.vue'
import OfflineHistory from '../views/OfflineHistory.vue'
export default createRouter({
  history: window.__TAURI_INTERNALS__ ? createWebHashHistory() : createWebHistory(),
  routes: [
    { path: '/', redirect: '/running' },
    { path: '/running', component: RealTime, meta: { title: '实时对局' } },
    { path: '/rank', component: RankList, meta: { title: '战绩中心' } },
    { path: '/archive', component: OfflineHistory, meta: { title: '离线记录' } },
    { path: '/settings', component: SettingPage, meta: { title: '设置' } },
  ],
})
