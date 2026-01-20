import { defineConfig } from 'vite'
import copy from 'rollup-plugin-copy'

export default defineConfig({
  base: './',
  resolve: {
    alias: {
      'vue': 'https://unpkg.com/vue@3/dist/vue.esm-browser.prod.js'
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'esbuild',
    rollupOptions: {
      external: ['vue'],
      output: {
        paths: {
          vue: 'https://unpkg.com/vue@3/dist/vue.esm-browser.prod.js'
        }
      }
    }
  },
  plugins: [
    copy({
      targets: [
        { src: 'api/*', dest: 'dist/api' },
        { src: 'manifest.json', dest: 'dist' }
      ],
      hook: 'writeBundle'
    })
  ],
  server: {
    port: 7100,
    host: true,
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
})
