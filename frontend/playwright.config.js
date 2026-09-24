import { defineConfig } from '@playwright/test'
export default defineConfig({
  testDir: './e2e',
  workers: 1,
  timeout: 120000,
  expect: { timeout: 15000 },
  use: {
    baseURL: 'http://localhost:5173',
    viewport: { width: 1200, height: 800 },
    trace: 'retain-on-failure',
    channel: process.env.PLAYWRIGHT_CHANNEL || undefined,
  },
  webServer: {
    command:
      'node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 5173 --strictPort',
    url: 'http://localhost:5173',
    env: {
      VITE_BACK_URL: 'http://127.0.0.1:9366',
      VITE_WS_URL: 'ws://127.0.0.1:9366/ws',
    },
    reuseExistingServer: false,
  },
})
