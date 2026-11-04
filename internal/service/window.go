package service

import (
	"strings"
	"unicode/utf8"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/rouarn/wails3-template/internal/contract"
)

const maxWindowTitleRunes = 80

// WindowService 把常用窗口操作收敛成受控接口，避免前端直接摸系统能力。
type WindowService struct {
	window application.Window
}

// NewWindowService 绑定主窗口。
func NewWindowService(window application.Window) *WindowService {
	return &WindowService{window: window}
}

// ServiceName 用于 Wails 日志中的服务标识。
func (s *WindowService) ServiceName() string {
	return "WindowService"
}

// Minimise 最小化主窗口。
func (s *WindowService) Minimise() {
	s.window.Minimise()
}

// ToggleMaximise 在最大化与还原之间切换，返回切换后的状态。
func (s *WindowService) ToggleMaximise() bool {
	if s.window.IsMaximised() {
		s.window.Restore()
	} else {
		s.window.Maximise()
	}
	return s.window.IsMaximised()
}

// Close 关闭主窗口（macOS 下等同退出）。
func (s *WindowService) Close() {
	s.window.Close()
}

// SetTitle 修改窗口标题，限制长度避免超长标题撑坏标题栏。
func (s *WindowService) SetTitle(title string) error {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return contract.NewInvalidArgument("标题不能为空", "")
	}
	if utf8.RuneCountInString(trimmed) > maxWindowTitleRunes {
		return contract.NewInvalidArgument("标题过长", "最多 80 个字符")
	}
	s.window.SetTitle(trimmed)
	return nil
}
