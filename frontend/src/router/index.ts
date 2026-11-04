import { createRouter, createWebHashHistory } from 'vue-router'

import BlankLayout from '../layouts/BlankLayout.vue'
import DefaultLayout from '../layouts/DefaultLayout.vue'

// Wails 的资源和路由都走自定义协议，使用 hash 模式避免服务端路径歧义。
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      component: DefaultLayout,
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('../views/HomeView.vue'),
          meta: { title: '首页' },
        },
        {
          path: 'capabilities',
          name: 'capabilities',
          component: () => import('../views/CapabilitiesView.vue'),
          meta: { title: '系统能力' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../views/SettingsView.vue'),
          meta: { title: '设置' },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      component: BlankLayout,
      children: [
        {
          path: '',
          name: 'not-found',
          component: () => import('../views/NotFoundView.vue'),
          meta: { title: '页面不存在' },
        },
      ],
    },
  ],
})

export default router
