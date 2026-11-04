import { Events } from '@wailsio/runtime'

import { AppClock, AppStatus } from '@bindings/events'

import type { ClockEvent, StatusEvent } from './dto'

// 事件订阅统一收口，返回取消函数供组件卸载时调用。
export function onClock(handler: (payload: ClockEvent) => void): () => void {
  return Events.On(AppClock, (event) => handler(event.data as ClockEvent))
}

export function onStatus(handler: (payload: StatusEvent) => void): () => void {
  return Events.On(AppStatus, (event) => handler(event.data as StatusEvent))
}
