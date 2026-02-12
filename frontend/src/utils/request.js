import axios from 'axios'
import { VueAxios } from './axios'

const backendBase = (import.meta.env.VITE_BACK_URL || '').replace(/\/$/, '')
const withBackendBase = (apiPath) => backendBase ? `${backendBase}${apiPath}` : apiPath

// 创建 axios 实例
const request = axios.create({
    // API 请求的默认前缀
    baseURL: withBackendBase('/v1')
    // baseURL: 'http://127.0.0.1:4523/m1/4930153-4587449-7bd90a27'
})

const requestMock = axios.create({
    baseURL: 'http://127.0.0.1:4523/m1/4930153-4587449-7bd90a27'
})

const localriotapi = axios.create({
    // API 请求的默认前缀
    baseURL: withBackendBase('/riot')
})

// 异常拦截处理器
const errorHandler = (error) => {
    if (error.response) {
        console.log(error.response)
    }
    return Promise.reject(error.response)
}


// response interceptor
request.interceptors.response.use((response) => {
    return response.data
}, errorHandler)


localriotapi.interceptors.response.use((response) => {
    return response.data
}, errorHandler)
requestMock.interceptors.response.use((response) => {
    return response.data
}, errorHandler)

const installer = {
    vm: {},
    install(Vue) {
        Vue.use(VueAxios, request, localriotapi)
    }
}

export default request

export {
    installer as VueAxios,
    request,
    localriotapi,
    requestMock
}
