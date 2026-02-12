<template>
    <a-layout style="min-height: 100vh">
        <a-layout-sider breakpoint="lg" :collapsible="false" :collapsed="false">
            <!-- <div class="logo" /> -->
            <a-card :bordered="false" class="custom-card" style="width: 200px">
                <template #cover>
                    <a-image :width="200" :height="200" :src="userInfo.icon" :preview="false"
                        fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMIAAADDCAYAAADQvc6UAAABRWlDQ1BJQ0MgUHJvZmlsZQAAKJFjYGASSSwoyGFhYGDIzSspCnJ3UoiIjFJgf8LAwSDCIMogwMCcmFxc4BgQ4ANUwgCjUcG3awyMIPqyLsis7PPOq3QdDFcvjV3jOD1boQVTPQrgSkktTgbSf4A4LbmgqISBgTEFyFYuLykAsTuAbJEioKOA7DkgdjqEvQHEToKwj4DVhAQ5A9k3gGyB5IxEoBmML4BsnSQk8XQkNtReEOBxcfXxUQg1Mjc0dyHgXNJBSWpFCYh2zi+oLMpMzyhRcASGUqqCZ16yno6CkYGRAQMDKMwhqj/fAIcloxgHQqxAjIHBEugw5sUIsSQpBobtQPdLciLEVJYzMPBHMDBsayhILEqEO4DxG0txmrERhM29nYGBddr//5/DGRjYNRkY/l7////39v///y4Dmn+LgeHANwDrkl1AuO+pmgAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAwqADAAQAAAABAAAAwwAAAAD9b/HnAAAHlklEQVR4Ae3dP3PTWBSGcbGzM6GCKqlIBRV0dHRJFarQ0eUT8LH4BnRU0NHR0UEFVdIlFRV7TzRksomPY8uykTk/zewQfKw/9znv4yvJynLv4uLiV2dBoDiBf4qP3/ARuCRABEFAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghggQAQZQKAnYEaQBAQaASKIAQJEkAEEegJmBElAoBEgghgg0Aj8i0JO4OzsrPv69Wv+hi2qPHr0qNvf39+iI97soRIh4f3z58/u7du3SXX7Xt7Z2enevHmzfQe+oSN2apSAPj09TSrb+XKI/f379+08+A0cNRE2ANkupk+ACNPvkSPcAAEibACyXUyfABGm3yNHuAECRNgAZLuYPgEirKlHu7u7XdyytGwHAd8jjNyng4OD7vnz51dbPT8/7z58+NB9+/bt6jU/TI+AGWHEnrx48eJ/EsSmHzx40L18+fLyzxF3ZVMjEyDCiEDjMYZZS5wiPXnyZFbJaxMhQIQRGzHvWR7XCyOCXsOmiDAi1HmPMMQjDpbpEiDCiL358eNHurW/5SnWdIBbXiDCiA38/Pnzrce2YyZ4//59F3ePLNMl4PbpiL2J0L979+7yDtHDhw8vtzzvdGnEXdvUigSIsCLAWavHp/+qM0BcXMd/q25n1vF57TYBp0a3mUzilePj4+7k5KSLb6gt6ydAhPUzXnoPR0dHl79WGTNCfBnn1uvSCJdegQhLI1vvCk+fPu2ePXt2tZOYEV6/fn31dz+shwAR1sP1cqvLntbEN9MxA9xcYjsxS1jWR4AIa2Ibzx0tc44fYX/16lV6NDFLXH+YL32jwiACRBiEbf5KcXoTIsQSpzXx4N28Ja4BQoK7rgXiydbHjx/P25TaQAJEGAguWy0+2Q8PD6/Ki4R8EVl+bzBOnZY95fq9rj9zAkTI2SxdidBHqG9+skdw43borCXO/ZcJdraPWdv22uIEiLA4q7nvvCug8WTqzQveOH26fodo7g6uFe/a17W3+nFBAkRYENRdb1vkkz1CH9cPsVy/jrhr27PqMYvENYNlHAIesRiBYwRy0V+8iXP8+/fvX11Mr7L7ECueb/r48eMqm7FuI2BGWDEG8cm+7G3NEOfmdcTQw4h9/55lhm7DekRYKQPZF2ArbXTAyu4kDYB2YxUzwg0gi/41ztHnfQG26HbGel/crVrm7tNY+/1btkOEAZ2M05r4FB7r9GbAIdxaZYrHdOsgJ/wCEQY0J74TmOKnbxxT9n3FgGGWWsVdowHtjt9Nnvf7yQM2aZU/TIAIAxrw6dOnAWtZZcoEnBpNuTuObWMEiLAx1HY0ZQJEmHJ3HNvGCBBhY6jtaMoEiJB0Z29vL6ls58vxPcO8/zfrdo5qvKO+d3Fx8Wu8zf1dW4p/cPzLly/dtv9Ts/EbcvGAHhHyfBIhZ6NSiIBTo0LNNtScABFyNiqFCBChULMNNSdAhJyNSiECRCjUbEPNCRAhZ6NSiAARCjXbUHMCRMjZqBQiQIRCzTbUnAARcjYqhQgQoVCzDTUnQIScjUohAkQo1GxDzQkQIWejUogAEQo121BzAkTI2agUIkCEQs021JwAEXI2KoUIEKFQsw01J0CEnI1KIQJEKNRsQ80JECFno1KIABEKNdtQcwJEyNmoFCJAhELNNtScABFyNiqFCBChULMNNSdAhJyNSiECRCjUbEPNCRAhZ6NSiAARCjXbUHMCRMjZqBQiQIRCzTbUnAARcjYqhQgQoVCzDTUnQIScjUohAkQo1GxDzQkQIWejUogAEQo121BzAkTI2agUIkCEQs021JwAEXI2KoUIEKFQsw01J0CEnI1KIQJEKNRsQ80JECFno1KIABEKNdtQcwJEyNmoFCJAhELNNtScABFyNiqFCBChULMNNSdAhJyNSiECRCjUbEPNCRAhZ6NSiAARCjXbUHMCRMjZqBQiQIRCzTbUnAARcjYqhQgQoVCzDTUnQIScjUohAkQo1GxDzQkQIWejUogAEQo121BzAkTI2agUIkCEQs021JwAEXI2KoUIEKFQsw01J0CEnI1KIQJEKNRsQ80JECFno1KIABEKNdtQcwJEyNmoFCJAhELNNtScABFyNiqFCBChULMNNSdAhJyNSiEC/wGgKKC4YMA4TAAAAABJRU5ErkJggg==" />
                </template>
                <a-card-meta :title="userInfo.name" class="custom-card-meta">
                    <template #description>
                        <p class="custom-description">召唤师等级：{{ userInfo.level }}</p>
                        <!-- <p class="custom-description">近20场胜率：52%</p> -->
                        <p class="custom-description">单双段位：{{ userInfo.rank }}</p>
                    </template>
                </a-card-meta>
            </a-card>
            <a-divider />
            <div class="account-center-team">
                <p>UID：{{ uid }} </p>
            </div>
        </a-layout-sider>
        <a-layout>
            <a-layout-header>
                <a-menu :selectedKeys="[$route.path]" theme="dark" mode="horizontal" :style="{ lineHeight: '64px' }">
                    <a-menu-item :disabled="true" style="width: 200px;">
                        <div>
                            状态：
                            <a-tag :color="status.color">
                                <template #icon>
                                    <sync-outlined :spin="true" v-if="status.color === 'processing'" />
                                    <check-circle-outlined v-if="status.color === 'success'" />
                                    <close-circle-outlined v-if="status.color === 'error'" />
                                </template>
                                {{ status.online }}
                            </a-tag>
                        </div>
                    </a-menu-item>
                    <a-menu-item key="/"><router-link to="/">基础配置</router-link></a-menu-item>
                    <a-menu-item key="/rank"><router-link to="/rank">战绩</router-link></a-menu-item>
                    <a-menu-item key="/running"><router-link to="/running">对局小助手</router-link></a-menu-item>
                </a-menu>
            </a-layout-header>
            <a-layout-content class="scale-content">
                <router-view />
            </a-layout-content>
            <a-layout-footer style="text-align: center;">
                lol-shield ©2022 Created by AnTengye
            </a-layout-footer>
        </a-layout>
    </a-layout>
</template>
<script>
import { defineComponent, ref, reactive, onMounted, watch } from 'vue';
import { getVersion, getUser, getSkins } from '@/api/bog'
import { useRoute, useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { createWebSocket } from './websocket/index'
import dicts from '@/model/dicts/index'
import {
    CheckCircleOutlined,
    SyncOutlined,
    CloseCircleOutlined,
} from '@ant-design/icons-vue';
export default defineComponent({
    components: {
        CheckCircleOutlined,
        SyncOutlined,
        CloseCircleOutlined,
    },
    setup() {
        const router = useRouter()
        const store = useStore()
        createWebSocket(store)
        const status = reactive({
            online: '服务暂未启动',
            color: 'error',
            version: 'v0.0.0'
        });
        const userInfo = reactive({
            name: '召唤师#99999',
            icon: '',
            level: 30,
            rank: '暂无数据'
        });
        const uid = ref(0)


        watch(() => store.getters['ws/getGameStatus'], (newStatus) => {
            if (newStatus === 2) {
                router.push('/running')
            }
        })
        watch(() => store.getters['ws/getSkinSync'], (newStatus) => {
            if (newStatus === 1) {
                let skins = localStorage.getItem('skins')
                if (skins === null || skins === {}) {
                    getSkins().then(res => {
                        localStorage.setItem('skins', JSON.stringify(res.data))
                    })
                }
            }
        })
        watch(() => store.getters['ws/getStatus'], (newOnline) => {
            if (newOnline === 1) {
                status.color = 'success'
                status.online = '已连接客户端'
                // 获取用户信息
                getUser().then(res => {
                    userInfo.icon = import.meta.env.VITE_BACK_URL + '/riot/v1/profile-icons/' + res.data.profileIconId + '.jpg'
                    userInfo.name = res.data.gameName + '#' + res.data.tagLine
                    userInfo.level = res.data.summonerLevel
                    if (res.data.tier !== 'NA') {
                        userInfo.rank = rank[res.data.tier] + res.data.division
                    }
                })

            } else if (newOnline === 2) {
                status.color = 'processing'
                status.online = '等待连接中'
            } else {
                status.color = 'error'
                status.online = '服务暂未启动'
            }
            console.log('status change', status)
        })
        watch(() => store.getters['ws/getUid'], (newUid) => {
            console.log('uid change', newUid)
            uid.value = newUid
        })
        const rank = dicts.getDict('rank')
        return {
            status,
            userInfo,
            // selectedKeys: ref(['1']),
            rank,
            uid
        };
    },
});

</script>
<style>
.site-layout-content {
    min-height: 280px;
    padding: 24px;
    background: #fff;
}

#components-layout-demo-top .logo {
    float: left;
    width: 120px;
    height: 31px;
    margin: 16px 24px 16px 0;
    background: rgba(255, 255, 255, 0.3);
}

.ant-row-rtl #components-layout-demo-top .logo {
    float: right;
    margin: 16px 0 16px 24px;
}

[data-theme="dark"] .site-layout-content {
    background: #141414;
}

.avatar-container {
    display: flex;
    justify-content: center;
    /* align-items: center; */
    padding-bottom: 20px;
}

.custom-card {
    background-color: transparent;
    /* 设置背景颜色为透明 */
    box-shadow: none;
    /* 可选：移除卡片的阴影 */
}

.custom-card-meta {
    color: #333;
    /* 设置 card-meta 的字体颜色 */
}

.custom-card-meta .ant-card-meta-title {
    color: #fcfcfc;
    /* 设置标题的字体颜色 */
}

.custom-description {
    color: #666;
    /* 设置描述的字体颜色 */
}

.account-center-team {
    position: absolute;
    bottom: 0;
    /* 固定在底部 */
    width: 100%;
    /* 占满父容器宽度 */
    text-align: center;
    /* 文本居中 */
    padding: 10px 0;
    /* 设置内边距 */
}

.scale-content {
    position: relative;
    /* padding-top: 56.25%; */
    /* 16:9 的比例 */
    /* min-width: 1920px; */
    /* min-height: 1080px; */
    width: calc(100%);
    /* height: 70vh; */
}

.scale-content>* {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
}
</style>