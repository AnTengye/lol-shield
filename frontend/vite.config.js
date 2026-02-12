import { defineConfig } from 'vite'
// import { UserConfig, ConfigEnv, loadEnv } from 'vite'
import { join } from 'path'
import vue from '@vitejs/plugin-vue'

function resolve(dir) {
    return join(__dirname, dir)
}
// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': resolve('src'),
        }
    },
    server: {
        proxy: {
            '/v1': {
                target: 'http://localhost:9365',
                changeOrigin: true,
            },
            '/riot': {
                target: 'http://localhost:9365',
                changeOrigin: true,
            },
            '/ws': {
                target: 'ws://localhost:9365',
                ws: true,
            }
        }
    }
})
