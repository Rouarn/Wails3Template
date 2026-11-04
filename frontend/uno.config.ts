import lucideIcons from '@iconify-json/lucide/icons.json'
import { defineConfig, presetIcons, presetUno, transformerDirectives } from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    // pnpm 隔离安装下 @iconify/utils 自动发现集合会失败，这里显式注入 lucide 集合；
    // presetIcons 仍只打包实际用到的图标。
    presetIcons({ scale: 1.2, collections: { lucide: () => lucideIcons } }),
  ],
  transformers: [transformerDirectives()],
  shortcuts: {
    'page-title': 'text-2xl font-600 leading-tight',
    'page-desc': 'text-sm opacity-70 mt-1',
    'card-shell': 'rounded-2xl border border-solid border-[var(--border-color)] bg-[var(--panel-bg)] p-5',
  },
})
