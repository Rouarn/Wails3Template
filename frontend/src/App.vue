<script setup lang="ts">
import { usePreferredDark } from '@vueuse/core'
import { dateEnUS, dateZhCN, darkTheme, enUS, zhCN } from 'naive-ui'
import { computed, watchEffect } from 'vue'

import { i18n } from './locales'
import { useAppStore } from './stores/app'

const store = useAppStore()
void store.bootstrap()

const preferredDark = usePreferredDark()
const isDark = computed(() => store.settings.theme === 'dark' || (store.settings.theme === 'system' && preferredDark.value))
const naiveTheme = computed(() => (isDark.value ? darkTheme : null))
const naiveLocale = computed(() => (store.settings.language === 'en-US' ? enUS : zhCN))
const naiveDateLocale = computed(() => (store.settings.language === 'en-US' ? dateEnUS : dateZhCN))

// data-theme 驱动 CSS 变量，保证非 Naive UI 的自定义外壳同步换肤。
watchEffect(() => {
  document.documentElement.dataset.theme = isDark.value ? 'dark' : 'light'
  i18n.global.locale.value = store.settings.language as 'zh-CN' | 'en-US'
})
</script>

<template>
  <n-config-provider :theme="naiveTheme" :locale="naiveLocale" :date-locale="naiveDateLocale">
    <router-view />
  </n-config-provider>
</template>
