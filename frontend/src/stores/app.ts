import { defineStore } from 'pinia'
import { ref } from 'vue'

import { getAppInfo, getSettings } from '../bridge/api'
import type { AppInfo, Settings } from '../bridge/dto'
import { toAppError, type AppErrorPayload } from '../bridge/errors'

export const useAppStore = defineStore('app', () => {
  const info = ref<AppInfo | null>(null)
  const settings = ref<Settings>({ theme: 'system', language: 'zh-CN', updatedAt: '' })
  const loading = ref(false)
  const error = ref<AppErrorPayload | null>(null)

  async function bootstrap(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [appInfo, appSettings] = await Promise.all([getAppInfo(), getSettings()])
      info.value = appInfo
      settings.value = appSettings
    } catch (err) {
      error.value = toAppError(err)
    } finally {
      loading.value = false
    }
  }

  function applySettings(next: Settings): void {
    settings.value = next
  }

  return { info, settings, loading, error, bootstrap, applySettings }
})
