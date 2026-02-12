import { createApp } from 'vue'
import Antd from 'ant-design-vue'
import App from './App.vue'
import store from './store'
import router from './router'
import { PageContainer } from '@ant-design-vue/pro-layout'
import 'ant-design-vue/dist/reset.css'
const app = createApp(App)

app.use(router).use(Antd).use(store).use(PageContainer).mount('#app')
