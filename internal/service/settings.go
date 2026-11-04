package service

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/rouarn/wails3-template/internal/contract"
	"github.com/rouarn/wails3-template/internal/infra"
)

// SettingsService 负责设置读写，并在保存成功后推送状态事件。
type SettingsService struct {
	app   *application.App
	store *infra.SettingsStore
}

// NewSettingsService 构建设置服务。
func NewSettingsService(app *application.App, store *infra.SettingsStore) *SettingsService {
	return &SettingsService{app: app, store: store}
}

// ServiceName 用于 Wails 日志中的服务标识。
func (s *SettingsService) ServiceName() string {
	return "SettingsService"
}

// Get 读取设置，首次启动返回默认值。
func (s *SettingsService) Get() (contract.Settings, error) {
	return s.store.Load()
}

// Save 校验枚举值后原子写入，并广播 settings-saved 状态事件。
func (s *SettingsService) Save(settings contract.Settings) error {
	if !contract.IsSupportedTheme(settings.Theme) {
		return contract.NewInvalidArgument("主题取值不受支持", settings.Theme)
	}
	if !contract.IsSupportedLanguage(settings.Language) {
		return contract.NewInvalidArgument("语言取值不受支持", settings.Language)
	}
	settings.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := s.store.Save(settings); err != nil {
		return fmt.Errorf("保存设置失败: %w", err)
	}
	s.emitStatus("settings-saved", "设置已保存")
	return nil
}

// Reset 恢复默认设置并落盘，便于设置页一键还原。
func (s *SettingsService) Reset() (contract.Settings, error) {
	settings := contract.DefaultSettings()
	settings.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := s.store.Save(settings); err != nil {
		return contract.Settings{}, fmt.Errorf("重置设置失败: %w", err)
	}
	s.emitStatus("settings-reset", "已恢复默认设置")
	return settings, nil
}

// 设置文件路径用于设置页展示，避免用户找不到落盘位置。
func (s *SettingsService) FilePath() string {
	return s.store.Path()
}

func (s *SettingsService) emitStatus(state string, message string) {
	now := time.Now().Format(time.RFC3339)
	s.app.Event.Emit(contract.EventStatus, contract.StatusEvent{State: state, Message: message, At: now})
}
