import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/feishu/',
  server: {
    port: 3001,
    host: true
  },
  build: {
    outDir: 'dist'
  },
  define: {
    // 解决飞书 SDK 使用 process.env 的问题
    'process.env': {}
  }
})
