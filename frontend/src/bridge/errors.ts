import { CancelError, type CancellablePromise } from '@wailsio/runtime'

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

// 后端通过 MarshalError 返回 {code,message,details}，运行时可能以对象或 JSON 字符串抛回，这里统一归一化。
export function toAppError(error: unknown): AppErrorPayload {
  if (typeof error === 'string') {
    try {
      return toAppError(JSON.parse(error))
    } catch {
      return { code: 'INTERNAL', message: error }
    }
  }
  // CancellablePromise.cancel() 会以 CancelError("Promise cancelled.") 拒绝，
  // 它不是业务错误；归一成 CANCELLED，避免调用方把它当成 INTERNAL 显示成红色报错。
  if (error instanceof CancelError) {
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
