import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/dingtalk/',
  server: {
    port: 3000,
    host: true
  },
  build: {
    outDir: 'dist',
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'antd': ['antd']
        }
      }
    }
  },
  // 抑制 Node 模块外部化警告（来自 @oclif/core 等依赖）
  optimizeDeps: {
    exclude: ['@oclif/core']
  }
})
