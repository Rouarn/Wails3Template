package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dirs 汇总各平台的落盘位置，避免路径散落在业务代码里。
type Dirs struct {
	AppName      string
	ConfigDir    string
	LogDir       string
	SettingsFile string
}

// NewDirs 按操作系统约定返回配置与日志目录：
// macOS: ~/Library/Application Support/<AppName>
// Linux: $XDG_CONFIG_HOME/<AppName>（默认 ~/.config）
// Windows: %AppData%\<AppName>
func NewDirs(appName string) (Dirs, error) {
	if appName == "" {
		return Dirs{}, fmt.Errorf("应用名不能为空")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return Dirs{}, fmt.Errorf("解析用户配置目录失败: %w", err)
	}
	configDir := filepath.Join(base, appName)
	logDir := filepath.Join(configDir, "logs")
	return Dirs{
		AppName:      appName,
		ConfigDir:    configDir,
		LogDir:       logDir,
		SettingsFile: filepath.Join(configDir, "settings.json"),
	}, nil
}
