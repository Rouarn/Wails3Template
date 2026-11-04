<script setup lang="ts">
import dayjs from 'dayjs'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { closeWindow, markReady, minimiseWindow, setWindowTitle, toggleMaximiseWindow } from '../bridge/api'
import type { ClockEvent, StatusEvent } from '../bridge/dto'
import { toAppError } from '../bridge/errors'
import { onClock, onStatus } from '../bridge/events'
import { useAppStore } from '../stores/app'

const { t } = useI18n()
const store = useAppStore()

const clock = ref<ClockEvent | null>(null)
const statusHistory = ref<StatusEvent[]>([])
const actionError = ref<string | null>(null)
const actionTip = ref<string | null>(null)
const titleDraft = ref('Wails v3 桌面模板')
const disposers: Array<() => void> = []

const latestStatus = computed(() => statusHistory.value[0] ?? null)
const clockText = computed(() => (clock.value ? dayjs(clock.value.now).format('HH:mm:ss') : t('home.waiting')))
const uptimeText = computed(() => (clock.value ? formatUptime(clock.value.uptimeSeconds) : '—'))

onMounted(() => {
  disposers.push(
    onClock((payload) => {
      clock.value = payload
    }),
    onStatus((payload) => {
      statusHistory.value = [payload, ...statusHistory.value].slice(0, 6)
    }),
  )
  // 订阅完成后再握手，确保 frontend-ready 一定被收到。
  markReady().catch(() => undefined)
})

onUnmounted(() => {
  disposers.forEach((dispose) => dispose())
})

function formatUptime(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  if (hours > 0) {
    return `${hours} 小时 ${minutes} 分 ${seconds} 秒`
  }
  if (minutes > 0) {
    return `${minutes} 分 ${seconds} 秒`
  }
  return `${seconds} 秒`
}

async function runAction(action: string, task: () => Promise<unknown>): Promise<void> {
  actionError.value = null
  actionTip.value = null
  try {
    await task()
    actionTip.value = `${action}成功`
  } catch (error) {
    actionError.value = toAppError(error).message
  }
}

async function handleToggleMaximise(): Promise<void> {
  actionError.value = null
  actionTip.value = null
  try {
    const maximised = await toggleMaximiseWindow()
    actionTip.value = maximised ? '窗口已最大化' : '窗口已还原'
  } catch (error) {
    actionError.value = toAppError(error).message
  }
}

async function handleSetTitle(): Promise<void> {
  if (titleDraft.value.trim().length === 0) {
    actionError.value = t('home.titleTooLong')
    return
  }
  await runAction('设置标题', () => setWindowTitle(titleDraft.value.trim()))
}
</script>

<template>
  <section class="flex flex-col gap-5">
    <header>
      <h1 class="page-title">{{ t('home.title') }}</h1>
      <p class="page-desc">{{ t('home.desc') }}</p>
    </header>

    <n-alert v-if="store.error" type="error" :title="store.error.message">
      {{ store.error.details }}
    </n-alert>
    <n-alert v-if="actionError" type="error" :title="actionError" closable @close="actionError = null" />
    <n-alert v-if="actionTip" type="success" :title="actionTip" closable @close="actionTip = null" />

    <n-card :title="t('home.environment')" size="small">
      <n-descriptions :column="2" label-placement="left" bordered>
        <n-descriptions-item :label="t('home.name')">
          {{ store.info?.name ?? t('common.loading') }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.version')">
          {{ store.info?.version ?? t('common.loading') }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.goVersion')">
          {{ store.info?.goVersion ?? '—' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.wailsVersion')">
          {{ store.info?.wailsVersion ?? '—' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.platform')">
          {{ store.info?.platform ?? '—' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.arch')">
          {{ store.info?.arch ?? '—' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.configDir')" :span="2">
          <span class="break-all">{{ store.info?.configDir ?? '—' }}</span>
        </n-descriptions-item>
        <n-descriptions-item :label="t('home.logDir')" :span="2">
          <span class="break-all">{{ store.info?.logDir ?? '—' }}</span>
        </n-descriptions-item>
      </n-descriptions>
    </n-card>

    <div class="grid grid-cols-1 gap-5 xl:grid-cols-2">
      <n-card :title="t('home.events')" size="small">
        <p class="page-desc mb-3">{{ t('home.eventsDesc') }}</p>
        <n-descriptions :column="1" label-placement="left" bordered>
          <n-descriptions-item :label="t('home.now')">
            <span class="live-value">{{ clockText }}</span>
          </n-descriptions-item>
          <n-descriptions-item :label="t('home.uptime')">
            <span class="live-value">{{ uptimeText }}</span>
          </n-descriptions-item>
          <n-descriptions-item :label="t('home.lastStatus')">
            {{ latestStatus?.message ?? t('home.waiting') }}
          </n-descriptions-item>
        </n-descriptions>

        <p class="page-desc mt-4 mb-2">{{ t('home.statusHistory') }}</p>
        <n-empty v-if="statusHistory.length === 0" :description="t('home.noStatus')" size="small" />
        <div v-else class="flex flex-col gap-2">
          <div v-for="item in statusHistory" :key="item.at + item.state" class="flex items-center gap-2 text-xs">
            <n-tag size="small" :bordered="false" type="info">{{ item.state }}</n-tag>
            <span class="opacity-80">{{ item.message }}</span>
            <span class="ml-auto opacity-50">{{ dayjs(item.at).format('HH:mm:ss') }}</span>
          </div>
        </div>
      </n-card>

      <n-card :title="t('home.window')" size="small">
        <p class="page-desc mb-3">{{ t('home.windowDesc') }}</p>
        <n-space vertical :size="12">
          <n-space>
            <n-button size="small" @click="runAction('最小化', minimiseWindow)">
              <template #icon>
                <span class="i-lucide-minus" />
              </template>
              {{ t('home.minimise') }}
            </n-button>
            <n-button size="small" @click="handleToggleMaximise">
              <template #icon>
                <span class="i-lucide-maximize" />
              </template>
              {{ t('home.maximise') }}
            </n-button>
            <n-button size="small" type="error" @click="runAction('关闭窗口', closeWindow)">
              <template #icon>
                <span class="i-lucide-x" />
              </template>
              {{ t('home.close') }}
            </n-button>
          </n-space>

          <n-divider class="!my-1" />

          <n-form-item :label="t('home.titleInput')" :show-feedback="false">
            <n-space align="center">
              <n-input
                v-model:value="titleDraft"
                maxlength="80"
                show-count
                :placeholder="t('home.titlePlaceholder')"
                style="width: 260px"
              />
              <n-button size="small" @click="handleSetTitle">{{ t('home.setTitle') }}</n-button>
            </n-space>
          </n-form-item>
        </n-space>
      </n-card>
    </div>
  </section>
</template>
