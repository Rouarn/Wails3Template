#!/usr/bin/env node
// 开发模式入口：挑选空闲端口 → 启动 wails3 dev → 退出时清理子进程。
import { spawn } from 'node:child_process'
import net from 'node:net'
import process from 'node:process'

const PORT_START = Number(process.env.WAILS_DEV_PORT ?? 9245)
const PORT_END = PORT_START + 20
const WAILS_BIN = process.env.WAILS3_BIN ?? 'wails3'

function isPortFree(port) {
  return new Promise((resolve) => {
    const server = net.createServer()
    server.once('error', () => resolve(false))
    server.once('listening', () => server.close(() => resolve(true)))
    server.listen(port, '127.0.0.1')
  })
}

async function pickPort() {
  for (let port = PORT_START; port <= PORT_END; port += 1) {
    if (await isPortFree(port)) {
      return port
    }
  }
  throw new Error(`端口 ${PORT_START}-${PORT_END} 全部被占用`)
}

const port = await pickPort()
console.log(`[dev] 使用端口 ${port}，执行 ${WAILS_BIN} dev`)

const child = spawn(WAILS_BIN, ['dev', '-config', './build/config.yml', '-port', String(port)], {
  stdio: 'inherit',
  env: { ...process.env, WAILS_VITE_PORT: String(port) },
})

let stopping = false
function shutdown(signal) {
  if (stopping) {
    return
  }
  stopping = true
  console.log(`\n[dev] 收到 ${signal}，正在退出…`)
  child.kill('SIGTERM')
  // 兜底：3 秒内仍未退出则强杀，避免残留 WebView 与 vite 进程。
  setTimeout(() => child.kill('SIGKILL'), 3000).unref()
}

for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => shutdown(signal))
}

child.on('error', (error) => {
  console.error(`[dev] 启动失败：${error.message}`)
  process.exit(1)
})

child.on('exit', (code) => {
  console.log(`[dev] 已退出（code=${code ?? 0}）`)
  process.exit(code ?? 0)
})
