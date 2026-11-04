#!/usr/bin/env node
// 最小回归验证：取消必须归一成 CANCELLED，且不能把真实错误吞成取消。
import assert from 'node:assert/strict'
import process from 'node:process'

import { CancelError, CancellablePromise, CancelledRejectionError } from '@wailsio/runtime'

import { run, toAppError } from '../src/bridge/errors.ts'

const failures = []

function expectCancel(name, error, expectedMessage) {
  const payload = toAppError(error)
  if (payload.code !== 'CANCELLED' || payload.message !== expectedMessage) {
    failures.push(`${name}: 期望 CANCELLED/${expectedMessage}，实际 ${payload.code}/${payload.message}`)
    return
  }
  console.log(`ok - ${name} -> ${payload.code}`)
}

function expectError(name, error, code, message) {
  const payload = toAppError(error)
  if (payload.code !== code || payload.message !== message) {
    failures.push(`${name}: 期望 ${code}/${message}，实际 ${payload.code}/${payload.message}`)
    return
  }
  console.log(`ok - ${name} -> ${payload.code}`)
}

// 1. 运行时取消的原始拒绝形态
expectCancel('CancelError 实例', new CancelError('Promise cancelled.'), '操作已取消')
expectCancel('字符串 Promise cancelled.', 'Promise cancelled.', '操作已取消')
expectCancel('DOMException AbortError', new DOMException('aborted', 'AbortError'), '操作已取消')
expectCancel('纯对象 CancelError（跨 realm 或结构化克隆）', { name: 'CancelError', message: 'Promise cancelled.' }, '操作已取消')
expectCancel('CancelledRejectionError', new CancelledRejectionError(Promise.resolve(), new Error('late rejection')), '操作已取消')

// 2. 真实错误不能被当成取消
expectError('普通 Error', new Error('boom'), 'INTERNAL', 'boom')
expectError('消息里带 cancel 的普通 Error', new Error('cancel failed'), 'INTERNAL', 'cancel failed')
expectError('无关字符串', 'network down', 'INTERNAL', 'network down')

// 3. Go 侧 marshal 出来的错误只认 code，且保留后端文案
expectError('Go CANCELLED', JSON.stringify({ code: 'CANCELLED', message: '任务已取消' }), 'CANCELLED', '任务已取消')
expectError('Go VALIDATION', JSON.stringify({ code: 'VALIDATION', message: '步数必须在 1-20 之间' }), 'VALIDATION', '步数必须在 1-20 之间')

// 4. run() 的真实链路：CancellablePromise 被取消后，调用方必须拿到 CANCELLED
const task = new CancellablePromise(() => {})
setTimeout(() => void task.cancel(), 20)
const rejection = await run(task, { action: '演示任务' }).then(
  () => null,
  (error) => error,
)
assert.equal(rejection?.code, 'CANCELLED', `run() 取消后应得到 CANCELLED，实际 ${rejection?.code}`)
console.log('ok - run() 取消链路 -> CANCELLED')

if (failures.length > 0) {
  console.error(`[cancelled] ${failures.length} 个用例失败：`)
  for (const failure of failures) {
    console.error(`  - ${failure}`)
  }
  process.exit(1)
}

console.log('[cancelled] 全部用例通过')
