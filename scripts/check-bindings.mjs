#!/usr/bin/env node
// 防漂移：重新生成 bindings，若与仓库中的产物有差异则失败。
import { execFileSync } from 'node:child_process'
import process from 'node:process'
import { abortWith, ensureWailsCli } from './lib/wails.mjs'

let cli
try {
  cli = ensureWailsCli()
} catch (error) {
  abortWith(error.message)
}

try {
  execFileSync(cli.bin, ['generate', 'bindings', '-clean=true', '-ts', '-i'], { stdio: 'inherit' })
} catch (error) {
  console.error(`[bindings] 生成失败：${error.message}`)
  process.exit(1)
}

const status = execFileSync('git', ['status', '--porcelain', '--', 'frontend/bindings'], { encoding: 'utf8' })
if (status.trim() !== '') {
  console.error('[bindings] 生成产物与仓库不一致，请提交更新后的 frontend/bindings：')
  console.error(status.trim())
  process.exit(1)
}

console.log('[bindings] 生成产物与仓库一致')
