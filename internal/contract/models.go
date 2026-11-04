package contract

// AppInfo 是运行信息页的数据来源。
type AppInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	GoVersion    string `json:"goVersion"`
	WailsVersion string `json:"wailsVersion"`
	Platform     string `json:"platform"`
	Arch         string `json:"arch"`
	ConfigDir    string `json:"configDir"`
	LogDir       string `json:"logDir"`
}

// Settings 是设置页持久化的用户偏好。
type Settings struct {
	Theme     string `json:"theme"`
	Language  string `json:"language"`
	UpdatedAt string `json:"updatedAt"`
}

// 主题与语言的允许值；前端下拉框与此保持一致。
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"

	LanguageZhCN = "zh-CN"
	LanguageEnUS = "en-US"
)

// DefaultSettings 返回首次启动时的默认设置。
func DefaultSettings() Settings {
	return Settings{
		Theme:    ThemeSystem,
		Language: LanguageZhCN,
	}
}

// IsSupportedTheme 判断主题是否在允许集合内。
func IsSupportedTheme(theme string) bool {
	switch theme {
	case ThemeSystem, ThemeLight, ThemeDark:
		return true
	default:
		return false
	}
}

// IsSupportedLanguage 判断语言是否在允许集合内。
func IsSupportedLanguage(language string) bool {
	switch language {
	case LanguageZhCN, LanguageEnUS:
		return true
	default:
		return false
	}
}
