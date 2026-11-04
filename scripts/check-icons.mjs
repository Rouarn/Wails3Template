#!/usr/bin/env node
// 图标守卫：UnoCSS 的 i-* 图标规则只在构建期扫描到字面量类名时才生成。
// 本脚本把「源码里用到的图标」与「产物 CSS 里真实存在的规则」对齐，
// 任何缺失都在构建期失败，避免出现「图标盒子占位还在、图标却空白」。
//
// 用法：
//   node scripts/check-icons.mjs            # 只校验图标名在 iconify 集合里真实存在
//   node scripts/check-icons.mjs --dist     # 额外校验 dist 产物 CSS 里每条规则都存在
import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const SRC_DIR = path.join(ROOT, 'packages/frontend/src')
const DIST_DIR = path.join(ROOT, 'packages/frontend/dist')
const ICONIFY_DIR = path.join(ROOT, 'packages/frontend/node_modules/@iconify-json')

const SCAN_EXTENSIONS = new Set(['.vue', '.ts', '.js', '.mjs', '.html', '.css'])
// 集合名不含连字符（lucide / mdi / carbon / tabler 都是一段）；图标名可以带连字符。
const ICON_CLASS = /\bi-([a-z0-9]+)-([a-z0-9]+(?:-[a-z0-9]+)*)\b/g
// 动态拼接：扫描器看不到具体类名，必须在 src/icons.ts 里登记字面量。
const DYNAMIC_PATTERNS = [
  { re: /\bi-[a-z0-9-]*\$\{/g, hint: '模板字符串拼接，类名在运行期才确定' },
  { re: /['"`]i-[a-z0-9-]*['"`]\s*\+/g, hint: '字符串拼接组装类名' },
]

const failures = []
const usages = []

function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      walk(full)
    } else if (SCAN_EXTENSIONS.has(path.extname(entry.name))) {
      scanFile(full)
    }
  }
}

function scanFile(file) {
  const text = fs.readFileSync(file, 'utf8')
  const rel = path.relative(ROOT, file)
  const lines = text.split('\n')

  lines.forEach((line, index) => {
    for (const { re, hint } of DYNAMIC_PATTERNS) {
      re.lastIndex = 0
      if (re.test(line)) {
        failures.push(`${rel}:${index + 1} 动态拼接图标类名（${hint}），扫描器无法生成对应 CSS 规则，请改用 src/icons.ts 的 APP_ICONS`)
      }
    }
    for (const match of line.matchAll(ICON_CLASS)) {
      usages.push({ collection: match[1], name: match[2], className: match[0], location: `${rel}:${index + 1}` })
    }
  })
}

function loadCollection(collection) {
  const file = path.join(ICONIFY_DIR, collection, 'icons.json')
  if (!fs.existsSync(file)) {
    return null
  }
  return JSON.parse(fs.readFileSync(file, 'utf8'))
}

const collectionCache = new Map()

function iconExists(collection, name) {
  if (!collectionCache.has(collection)) {
    collectionCache.set(collection, loadCollection(collection))
  }
  const data = collectionCache.get(collection)
  if (!data) {
    return false
  }
  return Boolean(data.icons?.[name] ?? data.aliases?.[name])
}

function readDistCss() {
  if (!fs.existsSync(DIST_DIR)) {
    return null
  }
  const assetsDir = path.join(DIST_DIR, 'assets')
  if (!fs.existsSync(assetsDir)) {
    return null
  }
  const cssFiles = fs.readdirSync(assetsDir).filter((name) => name.endsWith('.css'))
  if (cssFiles.length === 0) {
    return null
  }
  return cssFiles.map((name) => ({ name, css: fs.readFileSync(path.join(assetsDir, name), 'utf8') }))
}

function main() {
  walk(SRC_DIR)

  const unique = new Map()
  for (const usage of usages) {
    if (!unique.has(usage.className)) {
      unique.set(usage.className, usage)
    }
  }

  for (const usage of unique.values()) {
    if (!iconExists(usage.collection, usage.name)) {
      failures.push(`${usage.location} 图标 ${usage.className} 在 @iconify-json/${usage.collection} 集合里不存在（集合缺失或图标名拼错）`)
    }
  }

  let distMissing = []
  if (process.argv.includes('--dist')) {
    const dist = readDistCss()
    if (!dist) {
      failures.push('packages/frontend/dist 下没有 CSS 产物，请先执行 pnpm build:frontend')
    } else {
      distMissing = [...unique.keys()].filter((className) => !dist.some((file) => file.css.includes(`.${className}{`)))
      for (const className of distMissing) {
        const usage = unique.get(className)
        failures.push(
          `${usage.location} 图标 ${className} 未生成到产物 CSS（packages/frontend/dist/assets/*.css），UnoCSS 没扫到这个类名：请放进 src/icons.ts 的 APP_ICONS`,
        )
      }
    }
  }

  const collections = [...new Set([...unique.values()].map((u) => u.collection))].sort()
  console.log(`[icons] 扫描 ${path.relative(ROOT, SRC_DIR)}：${unique.size} 个图标类名，集合 ${collections.join(', ') || '无'}`)
  for (const usage of unique.values()) {
    console.log(`  - ${usage.className}  ${usage.location}`)
  }
  if (failures.length > 0) {
    console.error('[icons] 校验失败：')
    for (const failure of failures) {
      console.error(`  - ${failure}`)
    }
    process.exit(1)
  }
  console.log(`[icons] 通过${process.argv.includes('--dist') ? '（产物 CSS 均含对应规则）' : ''}`)
}

main()
