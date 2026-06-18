import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// В режиме разработки запросы /api проксируются на Go-бэкенд (localhost:8080).
// В продакшене (Docker) проксированием занимается nginx — см. nginx.conf.
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
