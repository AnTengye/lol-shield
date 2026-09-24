import axios from 'axios'
import { VueAxios } from './axios'
import { buildRuntimeBackendUrl } from './backend'

// 创建 axios 实例
const request = axios.create({
    // API 请求的默认前缀
    baseURL: buildRuntimeBackendUrl('/v1'),
        timeout: 15000
    // baseURL: 'http://127.0.0.1:4523/m1/4930153-4587449-7bd90a27'
})

const requestMock = axios.create({
    baseURL: 'http://127.0.0.1:4523/m1/4930153-4587449-7bd90a27'
})

const localriotapi = axios.create({
    // API 请求的默认前缀
    baseURL: buildRuntimeBackendUrl('/riot')
})

// 异常拦截处理器
const errorHandler = (error) => {
    const body = error.response?.data
    if (body?.message) error.message = body.field ? `${body.message}：${body.field}` : body.message
    if (body?.code) error.businessCode = body.code
    return Promise.reject(error)
}


// response interceptor
request.interceptors.response.use((response) => {
    if (response.data?.code != null && String(response.data.code) !== '0') {
        const error = new Error(response.data.message || '请求失败')
        error.businessCode = response.data.code
        return Promise.reject(error)
    }
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
