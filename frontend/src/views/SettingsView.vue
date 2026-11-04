<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { getSettings, getSettingsFilePath, resetSettings, saveSettings, writeClipboard } from '../bridge/api'
import type { Settings } from '../bridge/dto'
import { toAppError } from '../bridge/errors'
import { APP_ICONS } from '../icons'
import { useAppStore } from '../stores/app'

const { t } = useI18n()
const store = useAppStore()

const draft = reactive<Settings>({ ...store.settings })
const filePath = ref('')
const saving = ref(false)
const notice = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const copyState = ref<'idle' | 'ok' | 'error'>('idle')

// 主题选项带一张小预览色块，直接体现「跟随系统 / 浅色 / 深色」的观感差异。
const themeOptions = computed(() => [
  { label: t('settings.themeSystem'), value: 'system' },
  { label: t('settings.themeLight'), value: 'light' },
  { label: t('settings.themeDark'), value: 'dark' },
])

// 语言下拉显示各语言的原生写法，避免用户看不懂当前界面语言时无从选择。
const languageOptions = computed(() => [
  { label: t('settings.languageZh'), value: 'zh-CN' },
  { label: t('settings.languageEn'), value: 'en-US' },
])

// 外观与语言在草稿阶段就应用到全局，提供「点了马上看到」的预览；保存动作只负责落盘。
watch(
  () => [draft.theme, draft.language] as const,
  ([theme, language]) => {
    store.applySettings({ ...store.settings, theme, language })
  },
)

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

// 路径是只读信息，复制动作直接走 Go 侧剪贴板，避免浏览器权限差异。
async function handleCopyPath(): Promise<void> {
  if (!filePath.value) {
    return
  }
  try {
    const ok = await writeClipboard(filePath.value)
    copyState.value = ok ? 'ok' : 'error'
  } catch {
    copyState.value = 'error'
  }
  window.setTimeout(() => {
    copyState.value = 'idle'
  }, 2000)
}
</script>

<template>
  <section class="settings-page">
    <header>
      <h1 class="page-title">{{ t('settings.title') }}</h1>
      <p class="page-desc">{{ t('settings.desc') }}</p>
    </header>

    <n-alert v-if="notice" :type="notice.type" :title="notice.message" closable @close="notice = null" />

    <section class="settings-card">
      <header class="settings-card__head">
        <h2 class="settings-card__title">{{ t('settings.appearance') }}</h2>
        <p class="settings-card__desc">{{ t('settings.appearanceDesc') }}</p>
      </header>

      <div class="settings-row">
        <div class="settings-row__text">
          <p class="settings-row__title">{{ t('settings.theme') }}</p>
          <p class="settings-row__desc">{{ t('settings.themeDesc') }}</p>
        </div>
        <div class="settings-row__control">
          <div class="theme-picker" role="radiogroup" :aria-label="t('settings.theme')">
            <button
              v-for="option in themeOptions"
              :key="option.value"
              type="button"
              class="theme-option"
              role="radio"
              :aria-checked="draft.theme === option.value"
              :class="{ 'theme-option--active': draft.theme === option.value }"
              @click="draft.theme = option.value"
            >
              <span class="theme-swatch" :class="`theme-swatch--${option.value}`" aria-hidden="true">
                <span class="theme-swatch__line" />
                <span class="theme-swatch__line theme-swatch__line--short" />
              </span>
              <span class="theme-option__label">{{ option.label }}</span>
            </button>
          </div>
        </div>
      </div>
    </section>

    <section class="settings-card">
      <header class="settings-card__head">
        <h2 class="settings-card__title">{{ t('settings.language') }}</h2>
        <p class="settings-card__desc">{{ t('settings.languageDesc') }}</p>
      </header>

      <div class="settings-row">
        <div class="settings-row__text">
          <p class="settings-row__title">{{ t('settings.interfaceLanguage') }}</p>
          <p class="settings-row__desc">{{ t('settings.interfaceLanguageDesc') }}</p>
        </div>
        <div class="settings-row__control">
          <n-select v-model:value="draft.language" :options="languageOptions" class="settings-select" />
        </div>
      </div>
    </section>

    <section class="settings-card">
      <header class="settings-card__head">
        <h2 class="settings-card__title">{{ t('settings.storage') }}</h2>
        <p class="settings-card__desc">{{ t('settings.storageDesc') }}</p>
      </header>

      <div class="settings-row">
        <div class="settings-row__text">
          <p class="settings-row__title">{{ t('settings.filePath') }}</p>
          <p class="settings-row__desc">{{ t('settings.filePathDesc') }}</p>
        </div>
        <div class="settings-row__control settings-row__control--path">
          <code class="path-value">{{ filePath || '—' }}</code>
          <n-button size="small" secondary :disabled="!filePath" @click="handleCopyPath">
            <template #icon>
              <span :class="APP_ICONS.copy" class="text-sm" />
            </template>
            {{ copyState === 'ok' ? t('common.copied') : t('common.copy') }}
          </n-button>
          <span v-if="copyState === 'error'" class="copy-error">{{ t('settings.copyFailed') }}</span>
        </div>
      </div>

      <div class="settings-row">
        <div class="settings-row__text">
          <p class="settings-row__title">{{ t('settings.save') }}</p>
          <p class="settings-row__desc">{{ t('settings.desc') }}</p>
        </div>
        <div class="settings-row__control">
          <n-button type="primary" size="small" :loading="saving" :disabled="saving" @click="handleSave">
            {{ saving ? t('common.saving') : t('common.save') }}
          </n-button>
          <n-button size="small" :disabled="saving" @click="handleReset">{{ t('common.reset') }}</n-button>
        </div>
      </div>
    </section>

    <section class="settings-card">
      <header class="settings-card__head">
        <h2 class="settings-card__title">{{ t('settings.about') }}</h2>
        <p class="settings-card__desc">{{ t('settings.aboutDesc') }}</p>
      </header>

      <div class="settings-row">
        <p class="settings-row__title">{{ t('settings.appName') }}</p>
        <span class="settings-row__value">{{ store.info?.name ?? '—' }}</span>
      </div>
      <div class="settings-row">
        <p class="settings-row__title">{{ t('settings.appVersion') }}</p>
        <span class="settings-row__value">{{ store.info?.version ?? '—' }}</span>
      </div>
      <div class="settings-row">
        <p class="settings-row__title">{{ t('settings.goVersion') }}</p>
        <span class="settings-row__value">{{ store.info?.goVersion ?? '—' }}</span>
      </div>
      <div class="settings-row">
        <p class="settings-row__title">{{ t('settings.wailsVersion') }}</p>
        <span class="settings-row__value">{{ store.info?.wailsVersion ?? '—' }}</span>
      </div>
      <div class="settings-row">
        <p class="settings-row__title">{{ t('settings.platform') }}</p>
        <span class="settings-row__value">{{ store.info?.platform ?? '—' }} / {{ store.info?.arch ?? '—' }}</span>
      </div>
    </section>
  </section>
</template>
