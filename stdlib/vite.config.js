import { defineConfig } from 'vite';
import { resolve } from 'path';

// Get the library name from environment variable
const libName = process.env.LIB_NAME || 'lodash';

// Build configuration for a single stdlib bundle
export default defineConfig({
  build: {
    outDir: 'dist',
    emptyOutDir: false, // Don't clear on each build
    target: 'es2015',
    minify: true,
    lib: {
      entry: resolve(__dirname, `src/${libName}.js`),
      formats: ['iife'],
      name: libName,
      fileName: () => `${libName}.min.js`,
    },
    rollupOptions: {
      output: {
        // Wrap exports in module.exports for CommonJS compatibility
        footer: `if (typeof module !== 'undefined') { module.exports = ${libName}; }`,
      },
    },
  },
});
