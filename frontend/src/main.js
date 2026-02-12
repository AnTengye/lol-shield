import { createApp } from 'vue'
import Antd from 'ant-design-vue'
import App from './App.vue'
import store from './store'
import router from './router'
import { VueAxios } from './utils/request'
import { PageContainer } from '@ant-design-vue/pro-layout'
import 'ant-design-vue/dist/reset.css'
const app = createApp(App)

app.use(router).use(Antd).use(VueAxios).use(store).use(PageContainer).mount('#app')
