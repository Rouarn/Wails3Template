<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import brandStar from '../assets/brand-star.svg'
import { APP_ICONS } from '../icons'
import { useAppStore } from '../stores/app'

const { t } = useI18n()
const route = useRoute()
const store = useAppStore()

const navItems = [
  { name: 'home', to: '/', label: 'nav.home', icon: APP_ICONS.navHome },
  { name: 'capabilities', to: '/capabilities', label: 'nav.capabilities', icon: APP_ICONS.navCapabilities },
  { name: 'settings', to: '/settings', label: 'nav.settings', icon: APP_ICONS.navSettings },
] as const

// 激活项唯一来源：当前路由路径与菜单目标精确相等。
const isActive = (to: string) => route.path === to
</script>

<template>
  <div class="app-shell">
    <aside class="app-sidebar">
      <div class="mb-6 flex items-center gap-3 px-2">
        <div class="h-9 w-9 flex items-center justify-center rounded-xl bg-white/70">
          <img :src="brandStar" class="h-6 w-6" alt="" aria-hidden="true" />
        </div>
        <div>
          <p class="text-sm font-600 leading-tight">Wails v3 模板</p>
          <p class="text-xs leading-tight opacity-60">Vite Plus + Vue</p>
        </div>
      </div>

      <nav class="app-sidebar__nav flex flex-col gap-1">
        <router-link
          v-for="item in navItems"
          :key="item.name"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(item.to) }"
        >
          <span :class="item.icon" class="text-lg" aria-hidden="true" />
          <span>{{ t(item.label) }}</span>
        </router-link>
      </nav>

      <div class="app-sidebar__footer mt-auto px-2 pt-6 text-xs" :style="{ color: 'var(--text-muted)' }">
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
