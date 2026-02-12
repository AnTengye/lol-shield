import axios from 'axios'

const backendBase = (import.meta.env.VITE_BACK_URL || '').replace(/\/$/, '')
const withBackendBase = (apiPath) => backendBase ? `${backendBase}${apiPath}` : apiPath

export const requestV2 = axios.create({
    baseURL: withBackendBase('/api/v2')
})

requestV2.interceptors.response.use((response) => response.data, (error) => {
    if (error.response) {
        return Promise.reject(error.response.data)
    }
    return Promise.reject(error)
})
