import { openExternal } from './api'
import { toAppError } from './errors'

// v3 没有内置导航拦截：WebView 内任何 <a href> 跳转都由这里接管。
// 站内路由（#、/ 开头）交给 vue-router，外部链接一律走 Go 侧白名单。
export function installExternalLinkGuard(): void {
  document.addEventListener('click', (event) => {
    const target = event.target as HTMLElement | null
    const anchor = target?.closest('a[href]') as HTMLAnchorElement | null
    if (!anchor) {
      return
    }
    const href = anchor.getAttribute('href')?.trim() ?? ''
    if (href === '' || href.startsWith('#') || href.startsWith('/') || anchor.dataset.internal === 'true') {
      return
    }
    event.preventDefault()
    openExternal(href).catch((error) => {
      const payload = toAppError(error)
      console.warn(`[links] ${payload.code}: ${payload.message}`)
    })
  })
}
