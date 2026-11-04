import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import { defineConfig } from 'vite'
import wails from '@wailsio/runtime/plugins/vite'

// vite 通过包别名指向 @voidzero-dev/vite-plus-core，由 vp 命令行驱动构建。
export default defineConfig({
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@bindings': fileURLToPath(new URL('./bindings', import.meta.url)),
    },
  },
  server: {
    // 端口由 wails3 dev 或 scripts/dev.mjs 传入，禁止静默递增（strictPort）。
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [vue(), UnoCSS(), Components({ resolvers: [NaiveUiResolver()], dts: 'src/components.d.ts' }), wails('./bindings')],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
