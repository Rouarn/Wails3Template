import { CancelError, CancelledRejectionError, type CancellablePromise } from '@wailsio/runtime'

export interface AppErrorPayload {
  code: string
  message: string
  details?: string
}

interface RunOptions {
  action: string
  timeoutMs?: number
  cancel?: () => void
}

const defaultTimeoutMs = 15_000

// Wails v3 运行时主动取消时的拒绝形态（见 @wailsio/runtime 的 cancellable.js）：
// - CancelError，name = "CancelError"，message = "Promise cancelled."，cause 是 cancel() 的入参；
// - CancelledRejectionError，name = "CancelledRejectionError"，取消后才 reject 时用它包一层。
// DOM / AbortSignal 的取消统一是 AbortError；Go 侧 context 取消的原始文案是 "context canceled"。
// 这里只做「精确名字或整串消息」判断，避免把包含 cancel 字样的真实错误一并吞掉。
const cancellationNames = new Set(['CancelError', 'CancelledRejectionError', 'AbortError'])
const cancellationMessages = new Set(['Promise cancelled.', 'canceled', 'cancelled', 'context canceled', 'context cancelled'])
const cancelledRejectionPrefix = 'Unhandled rejection in cancelled promise.'

function isCancellation(error: unknown): boolean {
  if (error instanceof CancelError || error instanceof CancelledRejectionError) {
    return true
  }
  if (typeof error === 'string') {
    return cancellationMessages.has(error)
  }
  if (typeof error === 'object' && error !== null) {
    const record = error as Record<string, unknown>
    if (typeof record.name === 'string' && cancellationNames.has(record.name)) {
      return true
    }
    if (typeof record.message === 'string') {
      return cancellationMessages.has(record.message) || record.message.startsWith(cancelledRejectionPrefix)
    }
  }
  return false
}

// 后端通过 MarshalError 返回 {code,message,details}，运行时可能以对象或 JSON 字符串抛回，这里统一归一化。
export function toAppError(error: unknown): AppErrorPayload {
  if (typeof error === 'string') {
    if (isCancellation(error)) {
      return { code: 'CANCELLED', message: '操作已取消' }
    }
    try {
      return toAppError(JSON.parse(error))
    } catch {
      return { code: 'INTERNAL', message: error }
    }
  }
  if (isCancellation(error)) {
    return { code: 'CANCELLED', message: '操作已取消' }
  }
  if (error instanceof Error) {
    return { code: 'INTERNAL', message: error.message }
  }
  if (typeof error === 'object' && error !== null) {
    const record = error as Record<string, unknown>
    if (typeof record.code === 'string' && typeof record.message === 'string') {
      return {
        code: record.code,
        message: record.message,
        details: typeof record.details === 'string' ? record.details : undefined,
      }
    }
    if (record.error) {
      return toAppError(record.error)
    }
  }
  return { code: 'INTERNAL', message: '发生未预期的错误' }
}

// 统一补上超时与取消：绑定调用本身是可取消的 CancellablePromise。
export async function run<T>(task: CancellablePromise<T>, options: RunOptions): Promise<T> {
  const timeoutMs = options.timeoutMs ?? defaultTimeoutMs
  let timer: ReturnType<typeof setTimeout> | undefined
  try {
    return await Promise.race([
      task,
      new Promise<never>((_, reject) => {
        timer = setTimeout(() => {
          options.cancel?.()
          reject({ code: 'TIMEOUT', message: `${options.action}超时（${timeoutMs}ms）` })
        }, timeoutMs)
      }),
    ])
  } catch (error) {
    throw toAppError(error)
  } finally {
    if (timer) {
      clearTimeout(timer)
    }
  }
}
