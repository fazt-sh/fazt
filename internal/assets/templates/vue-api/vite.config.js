import { defineConfig } from 'vite'
import copy from 'rollup-plugin-copy'

export default defineConfig({
  base: './',
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'esbuild',
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
