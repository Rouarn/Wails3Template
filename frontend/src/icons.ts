// 全项目唯一的图标名来源。
//
// UnoCSS 的图标规则（`--un-icon` + mask）只在构建期扫描到 `i-<collection>-<name>`
// 字面量时才生成；模板里用变量绑定（`:class="item.icon"`）扫不到，产物 CSS 里就会
// 缺这条规则——表现为图标盒子还占着 18px 宽，图标本身却是空白。
//
// 因此：动态场景（数组字段、变量绑定）一律从这里取值，uno.config.ts 的 safelist
// 也直接由本表生成，不再手工维护。新增图标只改这一处，构建期校验脚本
// scripts/check-icons.mjs 会核对「集合里存在 + 产物 CSS 里有规则」。
export const APP_ICONS = {
  navHome: 'i-lucide-house',
  navSettings: 'i-lucide-settings',
  navCapabilities: 'i-lucide-sliders-horizontal',
  windowMinimise: 'i-lucide-minus',
  windowMaximise: 'i-lucide-maximize',
  windowClose: 'i-lucide-x',
  windowFullscreen: 'i-lucide-expand',
  windowPin: 'i-lucide-pin',
  copy: 'i-lucide-copy',
  folderOpen: 'i-lucide-folder-open',
  saveFile: 'i-lucide-save',
  bell: 'i-lucide-bell',
  clipboard: 'i-lucide-clipboard',
  externalLink: 'i-lucide-external-link',
  task: 'i-lucide-loader-circle',
} as const

export type AppIconName = keyof typeof APP_ICONS

// 交给 UnoCSS safelist：这些是变量绑定用到的类名，扫描器看不到，必须显式列出。
export const ICON_SAFELIST: string[] = Object.values(APP_ICONS)
