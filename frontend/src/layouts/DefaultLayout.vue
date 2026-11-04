<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { useAppStore } from '../stores/app'

const { t } = useI18n()
const store = useAppStore()

const navItems = [
  { name: 'home', to: '/', label: 'nav.home', icon: 'i-lucide-house' },
  { name: 'settings', to: '/settings', label: 'nav.settings', icon: 'i-lucide-settings' },
] as const
</script>

<template>
  <div class="app-shell">
    <aside class="app-sidebar">
      <div class="mb-6 flex items-center gap-3 px-2">
        <div class="h-9 w-9 flex items-center justify-center rounded-xl bg-indigo-500/15 text-indigo-400">
          <span class="i-lucide-layers text-xl" aria-hidden="true" />
        </div>
        <div>
          <p class="text-sm font-600 leading-tight">Wails v3 模板</p>
          <p class="text-xs leading-tight opacity-60">Vite Plus + Vue</p>
        </div>
      </div>

      <nav class="flex flex-col gap-1">
        <router-link v-for="item in navItems" :key="item.name" :to="item.to" class="nav-item">
          <span :class="item.icon" class="text-lg" aria-hidden="true" />
          <span>{{ t(item.label) }}</span>
        </router-link>
      </nav>

      <div class="mt-auto px-2 pt-6 text-xs" :style="{ color: 'var(--text-muted)' }">
        <p>版本 {{ store.info?.version ?? '—' }}</p>
        <p class="mt-1">Go 单进程 + 系统 WebView</p>
      </div>
    </aside>

    <main class="app-main">
      <div class="app-main__inner">
        <router-view />
      </div>
    </main>
  </div>
</template>
