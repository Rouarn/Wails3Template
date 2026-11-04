package service

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/rouarn/wails3-template/internal/contract"
	"github.com/rouarn/wails3-template/internal/platform"
)

// SystemService 提供运行信息与受限的外部链接打开能力。
type SystemService struct {
	app          *application.App
	dirs         platform.Dirs
	appVersion   string
	wailsVersion string
}

// NewSystemService 构造运行信息服务。
func NewSystemService(app *application.App, dirs platform.Dirs, appVersion string, wailsVersion string) *SystemService {
	return &SystemService{app: app, dirs: dirs, appVersion: appVersion, wailsVersion: wailsVersion}
}

// MarkReady 由前端在完成事件订阅后调用，回推一条 app.status，作为事件通道的就绪握手。
// 启动阶段直接 Emit 会早于前端订阅，因此改由前端触发。
func (s *SystemService) MarkReady() {
	s.app.Event.Emit(contract.EventStatus, contract.StatusEvent{
		State:   "frontend-ready",
		Message: "前端已挂载，事件通道就绪",
		At:      time.Now().Format(time.RFC3339),
	})
}

// ServiceName 用于 Wails 日志中的服务标识。
func (s *SystemService) ServiceName() string {
	return "SystemService"
}

// Info 返回运行信息页所需的全部字段。
func (s *SystemService) Info() contract.AppInfo {
	return contract.AppInfo{
		Name:         s.dirs.AppName,
		Version:      s.appVersion,
		GoVersion:    runtime.Version(),
		WailsVersion: s.wailsVersion,
		Platform:     runtime.GOOS,
		Arch:         runtime.GOARCH,
		ConfigDir:    s.dirs.ConfigDir,
		LogDir:       s.dirs.LogDir,
	}
}

// OpenExternal 只放行 http/https/mailto，其余协议一律拒绝，避免被前端当成任意协议启动器。
func (s *SystemService) OpenExternal(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return contract.NewInvalidArgument("链接不能为空", "OpenExternal 收到空字符串")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return contract.NewInvalidArgument("链接格式不合法", err.Error())
	}
	switch parsed.Scheme {
	case "http", "https":
		if parsed.Host == "" {
			return contract.NewInvalidArgument("链接缺少主机名", trimmed)
		}
	case "mailto":
		// mailto: 由系统邮件客户端接管，无需校验 Host。
	default:
		return contract.NewAppError(contract.CodeUnsupported, fmt.Sprintf("不允许打开 %q 协议的链接", parsed.Scheme))
	}
	if err := s.app.Browser.OpenURL(parsed.String()); err != nil {
		return fmt.Errorf("调用系统浏览器失败: %w", err)
	}
	return nil
}
