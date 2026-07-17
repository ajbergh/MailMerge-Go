import { defineConfig } from 'vitest/config';

// Note: the React fast-refresh plugin injects a preamble that fails under
// jsdom, so tests transform JSX via esbuild's automatic runtime instead.
export default defineConfig({
  esbuild: {
    jsx: 'automatic',
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
  },
});
