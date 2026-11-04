import { Events } from '@wailsio/runtime'

import { AppClock, AppNavigate, AppStatus, AppTaskProgress } from '@bindings/events'

import type { ClockEvent, NavigateEvent, StatusEvent, TaskProgressEvent } from './dto'

// 事件订阅统一收口，返回取消函数供组件卸载时调用。
export function onClock(handler: (payload: ClockEvent) => void): () => void {
  return Events.On(AppClock, (event) => handler(event.data as ClockEvent))
}

export function onStatus(handler: (payload: StatusEvent) => void): () => void {
  return Events.On(AppStatus, (event) => handler(event.data as StatusEvent))
}

// 后端任务的进度流：RunDemoTask 每推进一步就发一次。
export function onTaskProgress(handler: (payload: TaskProgressEvent) => void): () => void {
  return Events.On(AppTaskProgress, (event) => handler(event.data as TaskProgressEvent))
}

// 托盘「设置」等系统入口触发的路由跳转。
export function onNavigate(handler: (payload: NavigateEvent) => void): () => void {
  return Events.On(AppNavigate, (event) => handler(event.data as NavigateEvent))
}
