// Second build pass: bundles the standalone HIS collector as a plain script a
// third-party site can load with one <script> tag.
//
// It writes into the same dist/ the app build produces (emptyOutDir: false, so
// it must run after it), which is what `make frontend` copies into
// internal/dashboard/dist for go:embed. No extra embed or CI plumbing.
import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    // The app build owns cleaning dist/; this pass only adds a file to it.
    emptyOutDir: false,
    lib: {
      entry: 'src/lib/his-embed.ts',
      // IIFE, not ESM: integrators drop in a <script src>, without needing
      // type="module" or a bundler of their own.
      formats: ['iife'],
      name: 'GateCHAHIS',
      fileName: () => 'his.js',
    },
  },
})
