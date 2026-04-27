import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5116,
    proxy: {
      '/api': {
        target: 'http://localhost:8113',
        changeOrigin: true,
      },
    },
  },
})
