package app

import "net/http"

// 模板的安全基线：v3 没有内置 CSP 与导航拦截，这里用资源服务器中间件补上响应头。
// style-src 需要 'unsafe-inline'：Naive UI 与 UnoCSS 会注入行内样式。
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data:; " +
	"font-src 'self' data:; " +
	"connect-src 'self'; " +
	"frame-src 'none'; " +
	"object-src 'none'; " +
	"base-uri 'none'; " +
	"form-action 'none'"

// securityHeaders 给所有资源响应补安全头。
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Security-Policy", contentSecurityPolicy)
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
