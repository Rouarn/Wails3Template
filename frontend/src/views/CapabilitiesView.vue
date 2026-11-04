<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import {
  getSystemSummary,
  getWindowState,
  minimiseCapabilityWindow,
  notify,
  openExternalWithResult,
  pickDirectory,
  pickFiles,
  readClipboard,
  runDemoTask,
  saveDemoFile,
  toggleCapabilityAlwaysOnTop,
  toggleCapabilityFullscreen,
  toggleCapabilityMaximise,
  writeClipboard,
} from '../bridge/api'
import type { SelectedFiles, SystemSummary, TaskProgressEvent, WindowState } from '../bridge/dto'
import { toAppError } from '../bridge/errors'
import { onTaskProgress } from '../bridge/events'
import { APP_ICONS } from '../icons'

const { t } = useI18n()

type ResultKind = 'idle' | 'pending' | 'ok' | 'canceled' | 'error'

interface ActionState {
  kind: ResultKind
  message: string
}

// 每张卡片独立维护自己的执行状态，避免一个能力的失败影响其他卡片。
function createState(): ActionState {
  return { kind: 'idle', message: '' }
}

const fileState = reactive(createState())
const fileResult = ref<SelectedFiles | null>(null)
const directoryState = reactive(createState())
const directoryPath = ref('')
const saveState = reactive(createState())
const savePath = ref('')
const saveBytes = ref(0)

const notifyTitle = ref('Wails 系统能力')
const notifyBody = ref('这条通知由「系统能力」页面真实触发。')
const notifyState = reactive(createState())

const clipboardState = reactive(createState())
const clipboardText = ref('')

const linkInput = ref('https://www.wails.io/')
const linkState = reactive(createState())

const windowState = reactive(createState())
const windowInfo = ref<WindowState | null>(null)
const windowBusy = ref(false)

const taskState = reactive(createState())
const taskProgress = ref<TaskProgressEvent | null>(null)
const taskRunning = ref(false)
let taskCancel: (() => void) | null = null
let stopTaskProgress: (() => void) | null = null

const systemState = reactive(createState())
const summary = ref<SystemSummary | null>(null)

const windowChips = computed(() => {
  const current = windowInfo.value
  if (!current) {
    return []
  }
  return [
    { label: t('capabilities.stateMaximised'), value: current.maximised ? t('capabilities.yes') : t('capabilities.no') },
    { label: t('capabilities.stateFullscreen'), value: current.fullscreen ? t('capabilities.yes') : t('capabilities.no') },
    { label: t('capabilities.statePin'), value: current.alwaysOnTop ? t('capabilities.yes') : t('capabilities.no') },
    { label: t('capabilities.stateVisible'), value: current.visible ? t('capabilities.yes') : t('capabilities.no') },
    { label: t('capabilities.stateSize'), value: `${current.width} × ${current.height}` },
  ]
})

const platformLabel = computed(() => {
  const value = summary.value?.platform ?? ''
  switch (value) {
    case 'darwin':
      return 'macOS'
    case 'windows':
      return 'Windows'
    case 'linux':
      return 'Linux'
    default:
      return value || '—'
  }
})

const themeLabel = computed(() => {
  const value = summary.value?.theme ?? ''
  const labels: Record<string, string> = {
    system: t('settings.themeSystem'),
    light: t('settings.themeLight'),
    dark: t('settings.themeDark'),
  }
  return labels[value] ?? (value || '—')
})

const appearanceLabel = computed(() => {
  const value = summary.value?.systemAppearance ?? ''
  if (value === 'dark') {
    return t('settings.themeDark')
  }
  if (value === 'light') {
    return t('settings.themeLight')
  }
  return '—'
})

const screenLabel = computed(() => {
  const current = summary.value
  if (!current || current.screenWidth === 0) {
    return '—'
  }
  return `${current.screenWidth} × ${current.screenHeight}（缩放 ${current.screenScale.toFixed(2)}）`
})

onMounted(() => {
  void refreshWindowState()
  void refreshSystemSummary()
  stopTaskProgress = onTaskProgress((payload) => {
    taskProgress.value = payload
  })
})

onUnmounted(() => {
  stopTaskProgress?.()
  taskCancel?.()
})

function applyError(state: ActionState, error: unknown): void {
  state.kind = 'error'
  state.message = toAppError(error).message
}

async function handlePickFiles(): Promise<void> {
  fileState.kind = 'pending'
  fileState.message = ''
  fileResult.value = null
  try {
    const result = await pickFiles()
    if (result.canceled) {
      fileState.kind = 'canceled'
      fileState.message = t('capabilities.canceled')
      return
    }
    fileResult.value = result
    fileState.kind = 'ok'
    fileState.message = t('capabilities.fileCount', { count: result.count })
  } catch (error) {
    applyError(fileState, error)
  }
}

async function handlePickDirectory(): Promise<void> {
  directoryState.kind = 'pending'
  directoryState.message = ''
  directoryPath.value = ''
  try {
    const result = await pickDirectory()
    if (result.canceled) {
      directoryState.kind = 'canceled'
      directoryState.message = t('capabilities.canceled')
      return
    }
    directoryPath.value = result.path
    directoryState.kind = 'ok'
    directoryState.message = result.path
  } catch (error) {
    applyError(directoryState, error)
  }
}

async function handleSaveFile(): Promise<void> {
  saveState.kind = 'pending'
  saveState.message = ''
  savePath.value = ''
  saveBytes.value = 0
  try {
    const result = await saveDemoFile()
    if (result.canceled) {
      saveState.kind = 'canceled'
      saveState.message = t('capabilities.canceled')
      return
    }
    savePath.value = result.path
    saveBytes.value = result.bytes
    saveState.kind = 'ok'
    saveState.message = t('capabilities.savedBytes', { bytes: result.bytes })
  } catch (error) {
    applyError(saveState, error)
  }
}

async function handleNotify(): Promise<void> {
  notifyState.kind = 'pending'
  notifyState.message = ''
  try {
    const receipt = await notify(notifyTitle.value, notifyBody.value)
    notifyState.kind = receipt.delivered ? 'ok' : 'error'
    notifyState.message = receipt.message
  } catch (error) {
    applyError(notifyState, error)
  }
}

async function handleClipboardWrite(): Promise<void> {
  clipboardState.kind = 'pending'
  clipboardState.message = ''
  try {
    const ok = await writeClipboard('来自 Wails 系统能力页面的剪贴板示例文本')
    clipboardState.kind = ok ? 'ok' : 'error'
    clipboardState.message = ok ? t('capabilities.clipboardWritten') : t('common.failed')
  } catch (error) {
    applyError(clipboardState, error)
  }
}

async function handleClipboardRead(): Promise<void> {
  clipboardState.kind = 'pending'
  clipboardState.message = ''
  clipboardText.value = ''
  try {
    const result = await readClipboard()
    if (!result.available) {
      clipboardState.kind = 'canceled'
      clipboardState.message = t('capabilities.clipboardEmpty')
      return
    }
    clipboardText.value = result.text
    clipboardState.kind = 'ok'
    clipboardState.message = ''
  } catch (error) {
    applyError(clipboardState, error)
  }
}

async function handleOpenLink(): Promise<void> {
  linkState.kind = 'pending'
  linkState.message = ''
  try {
    await openExternalWithResult(linkInput.value)
    linkState.kind = 'ok'
    linkState.message = t('capabilities.linkOpened')
  } catch (error) {
    applyError(linkState, error)
  }
}

async function handleRunTask(): Promise<void> {
  taskState.kind = 'pending'
  taskState.message = ''
  taskProgress.value = null
  taskRunning.value = true
  const task = runDemoTask(10, 200)
  taskCancel = task.cancel
  try {
    const result = await task.result
    taskState.kind = 'ok'
    taskState.message = t('capabilities.taskDone', { steps: result.steps })
  } catch (error) {
    if (toAppError(error).code === 'CANCELLED') {
      taskState.kind = 'canceled'
      taskState.message = t('capabilities.taskCancelled')
    } else {
      applyError(taskState, error)
    }
  } finally {
    taskRunning.value = false
    taskCancel = null
  }
}

function handleCancelTask(): void {
  taskCancel?.()
}

async function runWindowAction(action: () => Promise<WindowState>): Promise<void> {
  windowBusy.value = true
  windowState.kind = 'pending'
  windowState.message = ''
  try {
    windowInfo.value = await action()
    windowState.kind = 'ok'
    windowState.message = ''
  } catch (error) {
    applyError(windowState, error)
  } finally {
    windowBusy.value = false
  }
}

async function refreshWindowState(): Promise<void> {
  try {
    windowInfo.value = await getWindowState()
  } catch (error) {
    applyError(windowState, error)
  }
}

async function refreshSystemSummary(): Promise<void> {
  systemState.kind = 'pending'
  systemState.message = ''
  try {
    summary.value = await getSystemSummary()
    systemState.kind = 'ok'
  } catch (error) {
    applyError(systemState, error)
  }
}
</script>

<template>
  <section class="capabilities-page">
    <header>
      <h1 class="page-title">{{ t('capabilities.title') }}</h1>
      <p class="page-desc">{{ t('capabilities.desc') }}</p>
    </header>

    <div class="cap-grid">
      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.folderOpen" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.file') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.fileDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" type="primary" data-cap="pick-files" :loading="fileState.kind === 'pending'" @click="handlePickFiles">
              {{ t('capabilities.pickFiles') }}
            </n-button>
          </div>
          <div v-if="fileState.kind === 'ok'" class="cap-result">
            <p class="cap-result__title">{{ fileState.message }}</p>
            <ul class="cap-list">
              <li v-for="path in fileResult?.paths ?? []" :key="path" class="cap-path">{{ path }}</li>
            </ul>
          </div>
          <p v-else-if="fileState.kind === 'canceled'" class="cap-note">{{ fileState.message }}</p>
          <p v-else-if="fileState.kind === 'error'" class="cap-error">{{ fileState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.folderOpen" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.directory') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.directoryDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" :loading="directoryState.kind === 'pending'" @click="handlePickDirectory">
              {{ t('capabilities.pickDirectory') }}
            </n-button>
          </div>
          <div v-if="directoryState.kind === 'ok'" class="cap-result">
            <code class="cap-path">{{ directoryPath }}</code>
          </div>
          <p v-else-if="directoryState.kind === 'canceled'" class="cap-note">{{ directoryState.message }}</p>
          <p v-else-if="directoryState.kind === 'error'" class="cap-error">{{ directoryState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.saveFile" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.save') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.saveDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" data-cap="save-file" :loading="saveState.kind === 'pending'" @click="handleSaveFile">
              {{ t('capabilities.saveFile') }}
            </n-button>
          </div>
          <div v-if="saveState.kind === 'ok'" class="cap-result">
            <p class="cap-result__title">{{ saveState.message }}</p>
            <code class="cap-path">{{ savePath }}</code>
          </div>
          <p v-else-if="saveState.kind === 'canceled'" class="cap-note">{{ saveState.message }}</p>
          <p v-else-if="saveState.kind === 'error'" class="cap-error">{{ saveState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.bell" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.notify') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.notifyDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <n-input v-model:value="notifyTitle" size="small" :placeholder="t('capabilities.notifyTitle')" />
          <n-input
            v-model:value="notifyBody"
            size="small"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 3 }"
            :placeholder="t('capabilities.notifyBody')"
          />
          <div class="cap-actions">
            <n-button size="small" type="primary" data-cap="notify" :loading="notifyState.kind === 'pending'" @click="handleNotify">
              {{ t('capabilities.notifySend') }}
            </n-button>
          </div>
          <div v-if="notifyState.kind === 'ok'" class="cap-result">
            <p class="cap-result__title">{{ notifyState.message }}</p>
          </div>
          <p v-else-if="notifyState.kind === 'error'" class="cap-error">{{ notifyState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.clipboard" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.clipboard') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.clipboardDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" :loading="clipboardState.kind === 'pending'" @click="handleClipboardWrite">
              {{ t('capabilities.clipboardWrite') }}
            </n-button>
            <n-button size="small" @click="handleClipboardRead">{{ t('capabilities.clipboardRead') }}</n-button>
          </div>
          <div v-if="clipboardState.kind === 'ok'" class="cap-result">
            <p v-if="clipboardState.message" class="cap-result__title">{{ clipboardState.message }}</p>
            <code v-if="clipboardText" class="cap-path">{{ clipboardText }}</code>
          </div>
          <p v-else-if="clipboardState.kind === 'canceled'" class="cap-note">{{ clipboardState.message }}</p>
          <p v-else-if="clipboardState.kind === 'error'" class="cap-error">{{ clipboardState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.externalLink" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.link') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.linkDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <n-input v-model:value="linkInput" size="small" :placeholder="t('capabilities.linkPlaceholder')" />
          <div class="cap-actions">
            <n-button size="small" data-cap="open-link" :loading="linkState.kind === 'pending'" @click="handleOpenLink">
              {{ t('capabilities.linkOpen') }}
            </n-button>
          </div>
          <p v-if="linkState.kind === 'error'" class="cap-error">{{ linkState.message }}</p>
          <p v-else-if="linkState.kind === 'ok'" class="cap-note">{{ linkState.message }}</p>
          <p class="cap-hint">{{ t('capabilities.linkBlockedHint') }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.task" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.task') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.taskDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" type="primary" data-cap="run-task" :loading="taskRunning" @click="handleRunTask">
              {{ t('capabilities.taskStart') }}
            </n-button>
            <n-button size="small" data-cap="cancel-task" :disabled="!taskRunning" @click="handleCancelTask">
              {{ t('capabilities.taskCancel') }}
            </n-button>
          </div>
          <n-progress
            type="line"
            :percentage="taskProgress?.percent ?? 0"
            :status="taskState.kind === 'error' ? 'error' : taskState.kind === 'ok' ? 'success' : 'default'"
          />
          <p class="cap-hint">{{ taskProgress?.message || taskState.message || t('capabilities.taskIdle') }}</p>
          <p v-if="taskState.kind === 'error'" class="cap-error">{{ taskState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.windowMaximise" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.window') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.windowDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <div class="cap-actions">
            <n-button size="small" :disabled="windowBusy" @click="runWindowAction(minimiseCapabilityWindow)">
              {{ t('capabilities.minimise') }}
            </n-button>
            <n-button size="small" :disabled="windowBusy" @click="runWindowAction(toggleCapabilityMaximise)">
              {{ t('capabilities.toggleMaximise') }}
            </n-button>
            <n-button size="small" :disabled="windowBusy" @click="runWindowAction(toggleCapabilityFullscreen)">
              {{ t('capabilities.toggleFullscreen') }}
            </n-button>
            <n-button size="small" :disabled="windowBusy" @click="runWindowAction(toggleCapabilityAlwaysOnTop)">
              {{ t('capabilities.togglePin') }}
            </n-button>
          </div>
          <div class="cap-chips">
            <span v-for="chip in windowChips" :key="chip.label" class="cap-chip">
              {{ chip.label }}<b>{{ chip.value }}</b>
            </span>
          </div>
          <p v-if="windowState.kind === 'error'" class="cap-error">{{ windowState.message }}</p>
        </div>
      </article>

      <article class="cap-card">
        <header class="cap-card__head">
          <span class="cap-card__icon"><span :class="APP_ICONS.navSettings" /></span>
          <div>
            <h2 class="cap-card__title">{{ t('capabilities.system') }}</h2>
            <p class="cap-card__desc">{{ t('capabilities.systemDesc') }}</p>
          </div>
        </header>
        <div class="cap-card__body">
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('capabilities.platform')">
              {{ platformLabel }}（{{ summary?.platform || '—' }}）
            </n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.arch')">{{ summary?.arch || '—' }}</n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.hostname')">
              {{ summary?.hostname || '—' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.screen')">{{ screenLabel }}</n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.language')">
              {{ summary?.language || '—' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.theme')">{{ themeLabel }}</n-descriptions-item>
            <n-descriptions-item :label="t('capabilities.appearance')">{{ appearanceLabel }}</n-descriptions-item>
          </n-descriptions>
          <div class="cap-actions">
            <n-button size="small" :loading="systemState.kind === 'pending'" @click="refreshSystemSummary">
              {{ t('capabilities.refresh') }}
            </n-button>
          </div>
          <p v-if="systemState.kind === 'error'" class="cap-error">{{ systemState.message }}</p>
        </div>
      </article>
    </div>
  </section>
</template>
