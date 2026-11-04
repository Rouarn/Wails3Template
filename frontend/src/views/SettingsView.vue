<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { getSettings, getSettingsFilePath, resetSettings, saveSettings } from '../bridge/api'
import type { Settings } from '../bridge/dto'
import { toAppError } from '../bridge/errors'
import { useAppStore } from '../stores/app'

const { t } = useI18n()
const store = useAppStore()

const draft = reactive<Settings>({ ...store.settings })
const filePath = ref('')
const saving = ref(false)
const notice = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const themeOptions = computed(() => [
  { label: t('settings.themeSystem'), value: 'system' },
  { label: t('settings.themeLight'), value: 'light' },
  { label: t('settings.themeDark'), value: 'dark' },
])

const languageOptions = computed(() => [
  { label: t('settings.languageZh'), value: 'zh-CN' },
  { label: t('settings.languageEn'), value: 'en-US' },
])

onMounted(async () => {
  try {
    const [current, path] = await Promise.all([getSettings(), getSettingsFilePath()])
    Object.assign(draft, current)
    store.applySettings(current)
    filePath.value = path
  } catch (error) {
    notice.value = { type: 'error', message: `${t('settings.loadFailed')}：${toAppError(error).message}` }
  }
})

async function handleSave(): Promise<void> {
  saving.value = true
  notice.value = null
  try {
    await saveSettings({ theme: draft.theme, language: draft.language, updatedAt: '' })
    const latest = await getSettings()
    Object.assign(draft, latest)
    store.applySettings(latest)
    notice.value = { type: 'success', message: t('settings.saved') }
  } catch (error) {
    notice.value = { type: 'error', message: `${t('settings.saveFailed')}：${toAppError(error).message}` }
  } finally {
    saving.value = false
  }
}

async function handleReset(): Promise<void> {
  saving.value = true
  notice.value = null
  try {
    const defaults = await resetSettings()
    Object.assign(draft, defaults)
    store.applySettings(defaults)
    notice.value = { type: 'success', message: t('settings.resetDone') }
  } catch (error) {
    notice.value = { type: 'error', message: `${t('settings.saveFailed')}：${toAppError(error).message}` }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="flex flex-col gap-5">
    <header>
      <h1 class="page-title">{{ t('settings.title') }}</h1>
      <p class="page-desc">{{ t('settings.desc') }}</p>
    </header>

    <n-alert v-if="notice" :type="notice.type" :title="notice.message" closable @close="notice = null" />

    <n-card size="small">
      <n-form label-placement="left" label-width="96" :show-feedback="false">
        <n-form-item :label="t('settings.theme')">
          <n-select v-model:value="draft.theme" :options="themeOptions" style="width: 240px" />
        </n-form-item>
        <n-form-item :label="t('settings.language')">
          <n-select v-model:value="draft.language" :options="languageOptions" style="width: 240px" />
        </n-form-item>
        <n-form-item :label="t('settings.filePath')">
          <span class="break-all text-xs opacity-70">{{ filePath || '—' }}</span>
        </n-form-item>
        <n-form-item :label="' '">
          <n-space>
            <n-button type="primary" :loading="saving" :disabled="saving" @click="handleSave">
              {{ saving ? t('common.saving') : t('common.save') }}
            </n-button>
            <n-button :disabled="saving" @click="handleReset">{{ t('common.reset') }}</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>
  </section>
</template>
