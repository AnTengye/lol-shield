import { createRouter, createWebHistory } from 'vue-router';
import SettingPage from '../views/SettingFrom.vue';
import RankList from '../views/RankList.vue';
import RealTime from '../views/RealTime.vue';

export default createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'index',
            meta: { title: 'Shield' },
            children: [
                {
                    path: '/',
                    name: 'Shield',
                    meta: { title: '配置', icon: 'icon-icon-test' },
                    component: SettingPage,
                },
                {
                    path: '/rank',
                    name: 'Rank',
                    meta: { title: '战绩', icon: 'icon-icon-test' },
                    component: RankList,
                },
                {
                    path: '/running',
                    name: 'RealTime',
                    meta: { title: '实时对局', icon: 'icon-icon-test' },
                    component: RealTime,
                },
            ],
        },
    ],
});
