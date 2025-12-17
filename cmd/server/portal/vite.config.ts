import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  base: '/portal/', // Set the base path to /portal/
  build: {
    outDir: '../static/portal', // Output to the static directory where Go server can access
    emptyOutDir: true,
    sourcemap: false
  },
  server: {
    proxy: {
      // Proxy all /api requests to the Go backend
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  }
})
