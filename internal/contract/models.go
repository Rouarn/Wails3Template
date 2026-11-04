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

// SelectedFiles 是文件选择对话框的返回值；Canceled 用于把「用户取消」与「调用失败」分开。
type SelectedFiles struct {
	Canceled bool     `json:"canceled"`
	Paths    []string `json:"paths"`
	Count    int      `json:"count"`
}

// SelectedDirectory 是目录选择对话框的返回值。
type SelectedDirectory struct {
	Canceled bool   `json:"canceled"`
	Path     string `json:"path"`
}

// SavedFile 是保存对话框的返回值，Bytes 为实际写入的字节数。
type SavedFile struct {
	Canceled bool   `json:"canceled"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
}

// NotificationReceipt 是系统通知的投递回执。
type NotificationReceipt struct {
	Delivered  bool   `json:"delivered"`
	Authorized bool   `json:"authorized"`
	Message    string `json:"message"`
}

// ClipboardText 是剪贴板读取结果；Available 为 false 表示读不到文本内容。
type ClipboardText struct {
	Available bool   `json:"available"`
	Text      string `json:"text"`
}

// SystemSummary 是「系统信息」表格的数据来源。
type SystemSummary struct {
	Platform         string  `json:"platform"`
	Arch             string  `json:"arch"`
	Hostname         string  `json:"hostname"`
	ScreenWidth      int     `json:"screenWidth"`
	ScreenHeight     int     `json:"screenHeight"`
	ScreenScale      float64 `json:"screenScale"`
	Language         string  `json:"language"`
	Theme            string  `json:"theme"`
	SystemAppearance string  `json:"systemAppearance"`
}

// WindowState 是窗口控制卡片的状态回显。
type WindowState struct {
	Maximised   bool `json:"maximised"`
	Fullscreen  bool `json:"fullscreen"`
	AlwaysOnTop bool `json:"alwaysOnTop"`
	Visible     bool `json:"visible"`
	Width       int  `json:"width"`
	Height      int  `json:"height"`
}

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
