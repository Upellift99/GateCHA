/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag === 'altcha-widget',
        },
      },
    }),
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/healthz': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'happy-dom',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      reportsDirectory: './coverage',
    },
  },
})
