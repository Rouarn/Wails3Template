<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
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
const loadError = ref('')
const copyState = ref<'idle' | 'ok' | 'error'>('idle')

type SaveStatus = 'idle' | 'pending' | 'saved' | 'failed'

const saveStatus = ref<SaveStatus>('idle')
const saveError = ref('')
const savedAt = ref('')
const failedAction = ref<'save' | 'reset' | null>(null)
const resetting = ref(false)

// 自动落盘的串行队列：写入永远按加入顺序执行，不会出现后发请求先落地。
// draftSerial/persistedSerial 记录「草稿改到第几版 / 已落盘第几版」，
// 循环会一直写到两版一致为止，因此连续切换主题时最终落盘的一定是最新选择。
let hydrated = false
let suppressAutoSave = false
let draftSerial = 0
let persistedSerial = 0
let debounceTimer: number | undefined
let writeQueue: Promise<void> = Promise.resolve()

const DEBOUNCE_MS = 300

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

// 外观与语言在草稿阶段就应用到全局，提供「点了马上看到」的预览；
// 落盘由下面的自动保存负责，页面不再提供手动保存按钮。
watch(
  () => [draft.theme, draft.language] as const,
  ([theme, language]) => {
    store.applySettings({ ...store.settings, theme, language })
    if (!hydrated || suppressAutoSave) {
      return
    }
    draftSerial += 1
    saveStatus.value = 'pending'
    saveError.value = ''
    failedAction.value = null
    // 300ms 防抖：连续切换主题只写最后一版，避免每次点击都落一次盘。
    if (debounceTimer !== undefined) {
      window.clearTimeout(debounceTimer)
    }
    debounceTimer = window.setTimeout(() => {
      debounceTimer = undefined
      void enqueueWrite()
    }, DEBOUNCE_MS)
  },
)

// 把一次写入排进串行队列；队列里永远只有按序执行的写入，没有并发。
function enqueueWrite(): Promise<void> {
  writeQueue = writeQueue.then(runWrite, runWrite)
  return writeQueue
}

// 一直写到「已落盘版本 == 草稿版本」为止；期间草稿又变了就继续写最新值。
async function runWrite(): Promise<void> {
  while (persistedSerial < draftSerial) {
    const serial = draftSerial
    const target: Settings = { theme: draft.theme, language: draft.language, updatedAt: '' }
    try {
      await saveSettings(target)
    } catch (error) {
      // 失败时不冒充成功：保留错误与重试入口，状态停留为 failed。
      saveStatus.value = 'failed'
      saveError.value = toAppError(error).message
      failedAction.value = 'save'
      return
    }
    persistedSerial = serial
    if (serial === draftSerial) {
      savedAt.value = formatClock(new Date())
      saveStatus.value = 'saved'
      saveError.value = ''
    }
  }
}

function formatClock(date: Date): string {
  return date.toLocaleTimeString(undefined, { hour12: false })
}

const saveStatusText = computed(() => {
  switch (saveStatus.value) {
    case 'pending':
      return t('settings.autoSavePending')
    case 'saved':
      return t('settings.autoSaveSaved', { time: savedAt.value })
    case 'failed':
      return `${t('settings.autoSaveFailed')}：${saveError.value}`
    default:
      return t('settings.autoSaveIdle')
  }
})

function handleRetry(): void {
  if (failedAction.value === 'reset') {
    void handleReset()
    return
  }
  saveStatus.value = 'pending'
  saveError.value = ''
  failedAction.value = null
  void enqueueWrite()
}

// 离开页面前把还在防抖窗口里的改动补写一次，避免快速切页丢掉最后一次修改。
onBeforeUnmount(() => {
  if (debounceTimer !== undefined) {
    window.clearTimeout(debounceTimer)
    debounceTimer = undefined
  }
  if (persistedSerial < draftSerial) {
    void enqueueWrite()
  }
})

onMounted(async () => {
  try {
    const [current, path] = await Promise.all([getSettings(), getSettingsFilePath()])
    Object.assign(draft, current)
    store.applySettings(current)
    filePath.value = path
  } catch (error) {
    loadError.value = toAppError(error).message
  } finally {
    // 等首次赋值的 watch 回调跑完再放行自动保存，避免页面一打开就多写一次盘。
    await nextTick()
    hydrated = true
  }
})

// 恢复默认：放进同一条串行队列，先等在途写入完成再由默认值覆盖，
// 顺序不会反过来（否则旧值可能盖掉重置结果）。
async function handleReset(): Promise<void> {
  if (resetting.value) {
    return
  }
  resetting.value = true
  suppressAutoSave = true
  if (debounceTimer !== undefined) {
    window.clearTimeout(debounceTimer)
    debounceTimer = undefined
  }
  const task = async (): Promise<void> => {
    const defaults = await resetSettings()
    draftSerial += 1
    Object.assign(draft, defaults)
    store.applySettings(defaults)
    persistedSerial = draftSerial
    await nextTick()
    savedAt.value = formatClock(new Date())
    saveStatus.value = 'saved'
    saveError.value = ''
    failedAction.value = null
  }
  writeQueue = writeQueue.then(task, task)
  try {
    await writeQueue
  } catch (error) {
    saveStatus.value = 'failed'
    saveError.value = toAppError(error).message
    failedAction.value = 'reset'
  } finally {
    suppressAutoSave = false
    resetting.value = false
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

    <n-alert v-if="loadError" type="error" :title="`${t('settings.loadFailed')}：${loadError}`" closable @close="loadError = ''" />

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
          <p class="settings-row__title">{{ t('settings.autoSave') }}</p>
          <p class="settings-row__desc">{{ t('settings.autoSaveDesc') }}</p>
        </div>
        <div class="settings-row__control">
          <span class="save-status" :class="`save-status--${saveStatus}`">{{ saveStatusText }}</span>
          <n-button v-if="saveStatus === 'failed'" size="small" @click="handleRetry">
            {{ t('settings.retry') }}
          </n-button>
        </div>
      </div>

      <div class="settings-row">
        <div class="settings-row__text">
          <p class="settings-row__title">{{ t('settings.resetTitle') }}</p>
          <p class="settings-row__desc">{{ t('settings.resetDesc') }}</p>
        </div>
        <div class="settings-row__control">
          <n-popconfirm
            :positive-text="t('settings.resetConfirmOk')"
            :negative-text="t('settings.resetConfirmCancel')"
            @positive-click="handleReset"
          >
            <template #trigger>
              <n-button size="small" :loading="resetting" :disabled="resetting">
                {{ t('common.reset') }}
              </n-button>
            </template>
            <p class="confirm-title">{{ t('settings.resetConfirmTitle') }}</p>
            <p class="confirm-body">{{ t('settings.resetConfirmBody') }}</p>
          </n-popconfirm>
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
