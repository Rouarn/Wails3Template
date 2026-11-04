#!/usr/bin/env node
// 通用透传入口：校验 CLI 版本后执行 wails3 子命令（build / package 等）。
// 用法：node scripts/wails.mjs build  |  node scripts/wails.mjs check
import { spawn } from 'node:child_process'
import process from 'node:process'
import { abortWith, ensureWailsCli } from './lib/wails.mjs'

const args = process.argv.slice(2)

let cli
try {
  cli = ensureWailsCli()
} catch (error) {
  abortWith(error.message)
}

if (args[0] === 'check') {
  console.log(`[cli] ${cli.bin} ${cli.version}（要求 >= v3.1.0）`)
  process.exit(0)
}

if (args.length === 0) {
  abortWith('[cli] 用法：node scripts/wails.mjs <wails3 子命令与参数>')
}

const child = spawn(cli.bin, args, { stdio: 'inherit', env: process.env })
child.on('error', (error) => {
  console.error(`[cli] 启动失败：${error.message}`)
  process.exit(1)
})
child.on('exit', (code, signal) => {
  process.exit(signal ? 1 : (code ?? 0))
})
