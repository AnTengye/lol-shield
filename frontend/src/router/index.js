import { createRouter, createWebHistory } from 'vue-router';

const SettingPage = () => import('../views/SettingFrom.vue');
const RealTime = () => import('../views/RealTime.vue');
const V2Console = () => import('../views/V2Console.vue');

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
                    path: '/running',
                    name: 'RealTime',
                    meta: { title: '实时对局', icon: 'icon-icon-test' },
                    component: RealTime,
                },
                {
                    path: '/v2',
                    name: 'V2Console',
                    meta: { title: 'V2控制台', icon: 'icon-icon-test' },
                    component: V2Console,
                },
            ],
        },
    ],
});
