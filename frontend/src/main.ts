import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { installExternalLinkGuard } from './bridge/links'
import { i18n } from './locales'
import router from './router'
import './styles/main.css'
import 'virtual:uno.css'

installExternalLinkGuard()

createApp(App).use(createPinia()).use(router).use(i18n).mount('#app')
