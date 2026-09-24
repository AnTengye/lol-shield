import { createApp } from 'vue'
import Antd from 'ant-design-vue'
import App from './App.vue'
import store from './store'
import router from './router'
import { VueAxios } from './utils/request'
import 'ant-design-vue/dist/reset.css'
import './style.css'
const app = createApp(App)

app.use(router).use(Antd).use(VueAxios).use(store).mount('#app')
