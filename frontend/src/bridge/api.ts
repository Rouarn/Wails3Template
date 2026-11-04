import type { CancellablePromise } from '@wailsio/runtime'

import * as CapabilitiesService from '@bindings/github.com/rouarn/wails3-template/internal/service/capabilitiesservice'
import * as SettingsService from '@bindings/github.com/rouarn/wails3-template/internal/service/settingsservice'
import * as SystemService from '@bindings/github.com/rouarn/wails3-template/internal/service/systemservice'
import * as WindowService from '@bindings/github.com/rouarn/wails3-template/internal/service/windowservice'

import type {
  AppInfo,
  ClipboardText,
  NotificationReceipt,
  SavedFile,
  SelectedDirectory,
  SelectedFiles,
  Settings,
  SystemSummary,
  WindowState,
} from './dto'
import { run } from './errors'

function invoke<T>(task: CancellablePromise<T>, action: string, timeoutMs?: number): Promise<T> {
  return run(task, { action, timeoutMs, cancel: () => void task.cancel() })
}

export function getAppInfo(): Promise<AppInfo> {
  return invoke(SystemService.Info(), '读取运行信息', 10_000)
}

// 外部链接只允许经由 Go 侧白名单打开，前端不做协议判断。
export function openExternal(url: string): Promise<void> {
  return invoke(SystemService.OpenExternal(url), '打开外部链接', 10_000)
}

// 前端完成事件订阅后回推一次就绪握手，后端据此推送 frontend-ready 状态。
export function markReady(): Promise<void> {
  return invoke(SystemService.MarkReady(), '事件通道握手', 5_000)
}

export function getSettings(): Promise<Settings> {
  return invoke(SettingsService.Get(), '读取设置', 10_000)
}

export function saveSettings(settings: Settings): Promise<void> {
  return invoke(SettingsService.Save(settings), '保存设置')
}

export function resetSettings(): Promise<Settings> {
  return invoke(SettingsService.Reset(), '重置设置')
}

export function getSettingsFilePath(): Promise<string> {
  return invoke(SettingsService.FilePath(), '读取设置文件路径', 10_000)
}

export function minimiseWindow(): Promise<void> {
  return invoke(WindowService.Minimise(), '最小化窗口', 5_000)
}

export function toggleMaximiseWindow(): Promise<boolean> {
  return invoke(WindowService.ToggleMaximise(), '切换窗口最大化', 5_000)
}

export function closeWindow(): Promise<void> {
  return invoke(WindowService.Close(), '关闭窗口', 5_000)
}

export function setWindowTitle(title: string): Promise<void> {
  return invoke(WindowService.SetTitle(title), '设置窗口标题', 5_000)
}

// 系统能力：对话框需要等待用户操作，超时给到 180s；通知可能弹出授权系统弹窗，给到 60s。
export function pickFiles(): Promise<SelectedFiles> {
  return invoke(CapabilitiesService.PickFiles(), '选择文件', 180_000)
}

export function pickDirectory(): Promise<SelectedDirectory> {
  return invoke(CapabilitiesService.PickDirectory(), '选择目录', 180_000)
}

export function saveDemoFile(): Promise<SavedFile> {
  return invoke(CapabilitiesService.SaveDemoFile(), '保存示例文件', 180_000)
}

export function notify(title: string, body: string): Promise<NotificationReceipt> {
  return invoke(CapabilitiesService.Notify(title, body), '发送系统通知', 60_000)
}

export function writeClipboard(text: string): Promise<boolean> {
  return invoke(CapabilitiesService.ClipboardWrite(text), '写入剪贴板', 5_000)
}

export function readClipboard(): Promise<ClipboardText> {
  return invoke(CapabilitiesService.ClipboardRead(), '读取剪贴板', 5_000)
}

export function openExternalWithResult(url: string): Promise<void> {
  return invoke(CapabilitiesService.OpenExternal(url), '打开外部链接', 10_000)
}

export function getWindowState(): Promise<WindowState> {
  return invoke(CapabilitiesService.WindowState(), '读取窗口状态', 5_000)
}

export function minimiseCapabilityWindow(): Promise<WindowState> {
  return invoke(CapabilitiesService.MinimiseWindow(), '最小化窗口', 5_000)
}

export function toggleCapabilityMaximise(): Promise<WindowState> {
  return invoke(CapabilitiesService.ToggleMaximiseWindow(), '切换窗口最大化', 5_000)
}

export function toggleCapabilityFullscreen(): Promise<WindowState> {
  return invoke(CapabilitiesService.ToggleFullscreenWindow(), '切换窗口全屏', 5_000)
}

export function toggleCapabilityAlwaysOnTop(): Promise<WindowState> {
  return invoke(CapabilitiesService.ToggleAlwaysOnTop(), '切换窗口置顶', 5_000)
}

export function getSystemSummary(): Promise<SystemSummary> {
  return invoke(CapabilitiesService.SystemSummary(), '读取系统信息', 10_000)
}
