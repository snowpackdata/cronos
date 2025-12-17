import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  base: '/admin/', // Set the base path to /admin/
  build: {
    outDir: '../static/admin', // Output to the static directory where Go server can access
    emptyOutDir: true,
    sourcemap: false,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: ['log'], // Remove console.log in production, keep console.error/warn
        drop_debugger: true
      }
    }
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
