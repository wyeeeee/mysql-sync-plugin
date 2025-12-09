import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/data/admin/',
  server: {
    port: 3001,
    proxy: {
      '/data/admin/api': {
        target: 'http://localhost:7139',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist'
  }
})
