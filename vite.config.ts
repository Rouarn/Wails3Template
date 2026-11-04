import { defineConfig } from 'vite-plus'

/**
 * 仓库级工具链配置：不参与构建，只被 `vp lint` / `vp fmt` / `vp check` 读取。
 * 生成产物（bindings、dist、bin）必须排除，否则会被反复格式化并与生成器打架。
 */
export default defineConfig({
  fmt: {
    semi: false,
    singleQuote: true,
    printWidth: 140,
    trailingComma: 'all',
    arrowParens: 'always',
    endOfLine: 'lf',
    // 生成产物、第三方平台资源与长文档都不参与格式化。
    ignorePatterns: [
      'frontend/bindings/**',
      'frontend/dist/**',
      'frontend/src/components.d.ts',
      'node_modules/**',
      '**/node_modules/**',
      'bin/**',
      'build/bin/**',
      // 平台打包模板含 Go template 占位符，格式化会破坏模板。
      'build/linux/nfpm/**',
      'build/windows/info.json',
      'build/windows/nsis/**',
      '.verify/**',
      'docs/**',
      '**/*.md',
    ],
  },
  lint: {
    ignorePatterns: [
      'dist/**',
      'frontend/dist/**',
      'frontend/bindings/**',
      'frontend/src/components.d.ts',
      'bin/**',
      'build/bin/**',
      'node_modules/**',
      '**/node_modules/**',
      '.verify/**',
      '*.min.js',
    ],
  },
  staged: {
    '*.{js,jsx,ts,tsx,mts,cts,vue,json,jsonc,md,yml,yaml}': 'vp check --fix',
  },
})
