package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"github.com/rouarn/wails3-template/internal/contract"
	"github.com/rouarn/wails3-template/internal/infra"
)

const (
	// 演示写入的文件名与内容长度都在前端回显，避免用户不知道写了什么。
	demoFileName = "wails3-capability-demo.txt"
	// 「所有文件」过滤器在 macOS/Windows 都可用。
	allFilesPattern = "*.*"
)

// 原生对话框的自定义文案集中在这里。
// 说明：Cancel / Where / Tags 这类系统级按钮来自 AppKit，不受这些参数控制，
// 只能靠应用包声明 zh-Hans 本地化（build/darwin/Info.plist 与 *.lproj）来中文化。
const (
	dialogPickFilesTitle   = "选择文件（可多选）"
	dialogPickFilesMessage = "可一次选择多个文件，支持文本、代码与图片"
	dialogPickFilesButton  = "选择"

	dialogPickDirectoryTitle   = "选择目录"
	dialogPickDirectoryMessage = "选择一个文件夹作为操作目标"
	dialogPickDirectoryButton  = "选择"

	dialogSaveFileTitle   = "保存示例文件"
	dialogSaveFileMessage = "将写入一段由系统能力页生成的示例文本"
	dialogSaveFileButton  = "保存"

	filterTextAndCode = "文本与代码"
	filterImages      = "图片"
	filterAllFiles    = "所有文件"
	filterTextFile    = "文本文件"
)

// CapabilitiesService 把系统能力收敛为一个个可演示、可回显的受控接口：
// 文件对话框、系统通知、剪贴板、外部链接、窗口控制与系统信息。
type CapabilitiesService struct {
	app      *application.App
	window   application.Window
	system   *SystemService
	store    *infra.SettingsStore
	notifier *notifications.NotificationService

	notifyOnce sync.Once
	notifyErr  error

	mu          sync.Mutex
	alwaysOnTop bool
}

// NewCapabilitiesService 构造系统能力服务；通知服务在此只创建实例，首次发送时才启动，
// 这样即使应用未被正确打包，也只在点击通知时给出明确错误，而不是阻塞启动。
func NewCapabilitiesService(
	app *application.App,
	window application.Window,
	system *SystemService,
	store *infra.SettingsStore,
	notifier *notifications.NotificationService,
) *CapabilitiesService {
	return &CapabilitiesService{app: app, window: window, system: system, store: store, notifier: notifier}
}

// ServiceName 用于 Wails 日志中的服务标识。
func (s *CapabilitiesService) ServiceName() string {
	return "CapabilitiesService"
}

// PickFiles 打开系统文件选择对话框（支持多选与扩展名过滤）。
// 用户取消对话框时返回 Canceled=true 且不返回错误，前端据此展示「已取消」。
func (s *CapabilitiesService) PickFiles() (contract.SelectedFiles, error) {
	dialog := s.app.Dialog.OpenFile()
	dialog.SetTitle(dialogPickFilesTitle)
	dialog.SetMessage(dialogPickFilesMessage)
	dialog.SetButtonText(dialogPickFilesButton)
	dialog.AddFilter(filterTextAndCode, "*.txt;*.md;*.json;*.go;*.ts;*.vue")
	dialog.AddFilter(filterImages, "*.png;*.jpg;*.jpeg")
	dialog.AddFilter(filterAllFiles, allFilesPattern)
	paths, err := dialog.PromptForMultipleSelection()
	if err != nil {
		return contract.SelectedFiles{}, fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if len(paths) == 0 {
		return contract.SelectedFiles{Canceled: true, Paths: []string{}}, nil
	}
	return contract.SelectedFiles{Paths: paths, Count: len(paths)}, nil
}

// PickDirectory 打开系统目录选择对话框；取消与失败同样分开返回。
func (s *CapabilitiesService) PickDirectory() (contract.SelectedDirectory, error) {
	dialog := s.app.Dialog.OpenFile()
	dialog.SetTitle(dialogPickDirectoryTitle)
	dialog.SetMessage(dialogPickDirectoryMessage)
	dialog.SetButtonText(dialogPickDirectoryButton)
	dialog.CanChooseFiles(false)
	dialog.CanChooseDirectories(true)
	dialog.CanCreateDirectories(true)
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return contract.SelectedDirectory{}, fmt.Errorf("打开目录对话框失败: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return contract.SelectedDirectory{Canceled: true}, nil
	}
	return contract.SelectedDirectory{Path: path}, nil
}

// SaveDemoFile 打开保存对话框并写入一段示例文本，返回实际落盘路径与字节数。
func (s *CapabilitiesService) SaveDemoFile() (contract.SavedFile, error) {
	// 保存对话框没有 SetTitle 方法，标题与过滤器通过 Options 传入。
	dialog := s.app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title:      dialogSaveFileTitle,
		Message:    dialogSaveFileMessage,
		ButtonText: dialogSaveFileButton,
		Filename:   demoFileName,
		Filters:    []application.FileFilter{{DisplayName: filterTextFile, Pattern: "*.txt"}},
	})
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return contract.SavedFile{}, fmt.Errorf("打开保存对话框失败: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return contract.SavedFile{Canceled: true}, nil
	}

	content := fmt.Sprintf(
		"这是由 Wails 系统能力页面触发的一次真实写盘。\n写入时间：%s\n目录：%s\n",
		time.Now().Format(time.RFC3339),
		path,
	)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return contract.SavedFile{}, fmt.Errorf("写入文件失败: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return contract.SavedFile{}, fmt.Errorf("读取写入结果失败: %w", err)
	}
	return contract.SavedFile{Path: path, Bytes: info.Size()}, nil
}

// Notify 发送一条系统通知。未授权时会先请求权限，仍被拒绝则返回 Authorized=false 的回执，
// 而不是抛错——「权限被拒」与「调用失败」是两回事，前端要给出不同提示。
func (s *CapabilitiesService) Notify(title string, body string) (contract.NotificationReceipt, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return contract.NotificationReceipt{}, contract.NewInvalidArgument("通知标题不能为空", "")
	}

	if err := s.startNotifier(); err != nil {
		return contract.NotificationReceipt{}, contract.NewAppError(contract.CodeUnsupported, fmt.Sprintf("系统通知不可用：%v", err))
	}

	authorized, err := s.notifier.CheckNotificationAuthorization()
	if err != nil {
		return contract.NotificationReceipt{}, fmt.Errorf("检查通知权限失败: %w", err)
	}
	if !authorized {
		authorized, err = s.notifier.RequestNotificationAuthorization()
		if err != nil {
			return contract.NotificationReceipt{}, fmt.Errorf("请求通知权限失败: %w", err)
		}
	}
	if !authorized {
		return contract.NotificationReceipt{
			Authorized: false,
			Message:    "系统通知权限未授予，请前往「系统设置 → 通知」允许本应用发送通知",
		}, nil
	}

	options := notifications.NotificationOptions{
		ID:       fmt.Sprintf("capability-demo-%d", time.Now().UnixNano()),
		Title:    trimmedTitle,
		Body:     strings.TrimSpace(body),
		Sound:    &notifications.NotificationSound{Silent: true},
		ThreadID: "wails3-template",
	}
	if err := s.notifier.SendNotification(options); err != nil {
		return contract.NotificationReceipt{}, fmt.Errorf("发送通知失败: %w", err)
	}
	return contract.NotificationReceipt{Authorized: true, Delivered: true, Message: "已提交给系统通知中心"}, nil
}

// ClipboardWrite 把文本写入系统剪贴板，返回写入是否成功。
func (s *CapabilitiesService) ClipboardWrite(text string) bool {
	return s.app.Clipboard.SetText(text)
}

// ClipboardRead 读取系统剪贴板文本；没有文本内容时 Available 为 false。
func (s *CapabilitiesService) ClipboardRead() contract.ClipboardText {
	text, ok := s.app.Clipboard.Text()
	if !ok {
		return contract.ClipboardText{}
	}
	return contract.ClipboardText{Available: true, Text: text}
}

// OpenExternal 复用 SystemService 的协议白名单（仅 http/https/mailto）。
func (s *CapabilitiesService) OpenExternal(rawURL string) error {
	return s.system.OpenExternal(rawURL)
}

// RunDemoTask 演示后端任务 + 事件流：按步推进并通过 app.task.progress 推送进度。
// 前端取消调用时 Wails 会取消 ctx，循环在下一步 select 时立即退出。
func (s *CapabilitiesService) RunDemoTask(ctx context.Context, steps int, delayMs int) (contract.TaskResult, error) {
	return runDemoTask(ctx, steps, delayMs, func(event contract.TaskProgressEvent) {
		s.app.Event.Emit(contract.EventTaskProgress, event)
	}, waitOrCancel)
}

// eventEmitter 与 waitOrCancel 抽出来便于单测替换，不影响生产路径。
type eventEmitter func(contract.TaskProgressEvent)

// waitOrCancel 等待指定时长；ctx 先结束则返回 false，表示任务被取消。
func waitOrCancel(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func runDemoTask(ctx context.Context, steps int, delayMs int, emit eventEmitter, wait func(context.Context, time.Duration) bool) (contract.TaskResult, error) {
	if steps < 1 || steps > 20 {
		return contract.TaskResult{}, contract.NewInvalidArgument("steps 超出范围", "允许 1–20")
	}
	if delayMs < 50 || delayMs > 2000 {
		return contract.TaskResult{}, contract.NewInvalidArgument("delayMs 超出范围", "允许 50–2000 毫秒")
	}
	for step := 1; step <= steps; step++ {
		if !wait(ctx, time.Duration(delayMs)*time.Millisecond) {
			emit(contract.TaskProgressEvent{
				Step: step - 1, Total: steps, Percent: (step - 1) * 100 / steps,
				State: "cancelled", Message: "任务已取消",
			})
			return contract.TaskResult{}, contract.NewAppError(contract.CodeCancelled, "任务已取消")
		}
		emit(contract.TaskProgressEvent{
			Step: step, Total: steps, Percent: step * 100 / steps,
			State: "running", Message: fmt.Sprintf("已完成 %d/%d", step, steps),
		})
	}
	return contract.TaskResult{Steps: steps, Completed: true}, nil
}

// MinimiseWindow 最小化主窗口并回传最新状态。
func (s *CapabilitiesService) MinimiseWindow() contract.WindowState {
	s.window.Minimise()
	return s.WindowState()
}

// ToggleMaximiseWindow 在最大化与还原之间切换并回传最新状态。
func (s *CapabilitiesService) ToggleMaximiseWindow() contract.WindowState {
	if s.window.IsMaximised() {
		s.window.Restore()
	} else {
		s.window.Maximise()
	}
	return s.WindowState()
}

// ToggleFullscreenWindow 在进入与退出全屏之间切换并回传最新状态。
func (s *CapabilitiesService) ToggleFullscreenWindow() contract.WindowState {
	if s.window.IsFullscreen() {
		s.window.UnFullscreen()
	} else {
		s.window.Fullscreen()
	}
	return s.WindowState()
}

// ToggleAlwaysOnTop 切换窗口置顶；Wails 暂未提供读取接口，因此状态由服务自身维护。
func (s *CapabilitiesService) ToggleAlwaysOnTop() contract.WindowState {
	s.mu.Lock()
	s.alwaysOnTop = !s.alwaysOnTop
	enabled := s.alwaysOnTop
	s.mu.Unlock()
	s.window.SetAlwaysOnTop(enabled)
	return s.WindowState()
}

// WindowState 回传窗口当前的尺寸、最大化/全屏/置顶/可见状态。
func (s *CapabilitiesService) WindowState() contract.WindowState {
	width, height := s.window.Size()
	s.mu.Lock()
	onTop := s.alwaysOnTop
	s.mu.Unlock()
	return contract.WindowState{
		Maximised:   s.window.IsMaximised(),
		Fullscreen:  s.window.IsFullscreen(),
		AlwaysOnTop: onTop,
		Visible:     s.window.IsVisible(),
		Width:       width,
		Height:      height,
	}
}

// SystemSummary 汇总平台、架构、主机名、主屏分辨率、系统语言与主题，供表格直接展示。
func (s *CapabilitiesService) SystemSummary() contract.SystemSummary {
	summary := contract.SystemSummary{
		Platform:         runtime.GOOS,
		Arch:             runtime.GOARCH,
		Language:         detectSystemLanguage(),
		SystemAppearance: detectSystemAppearance(),
	}
	if hostname, err := os.Hostname(); err == nil {
		summary.Hostname = hostname
	}
	if screen := s.app.Screen.GetPrimary(); screen != nil {
		summary.ScreenWidth = screen.Bounds.Width
		summary.ScreenHeight = screen.Bounds.Height
		summary.ScreenScale = float64(screen.ScaleFactor)
	}
	settings, err := s.store.Load()
	if err != nil {
		settings = contract.DefaultSettings()
	}
	summary.Theme = settings.Theme
	return summary
}

// startNotifier 延迟启动通知服务：服务启动时会校验运行环境是否支持通知。
func (s *CapabilitiesService) startNotifier() error {
	s.notifyOnce.Do(func() {
		s.notifyErr = s.notifier.ServiceStartup(context.Background(), application.ServiceOptions{})
	})
	return s.notifyErr
}

// detectSystemLanguage 依次读取 POSIX 语言变量；macOS 图形进程通常没有 LANG，
// 再回退到 AppleLocale（例如 zh_CN）。
func detectSystemLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("defaults", "read", "-g", "AppleLocale").Output(); err == nil {
			if value := strings.TrimSpace(string(output)); value != "" {
				return value
			}
		}
	}
	return "unknown"
}

// detectSystemAppearance 返回 dark / light / unknown，仅 macOS 有稳定的命令行读取方式。
func detectSystemAppearance() string {
	if runtime.GOOS != "darwin" {
		return "unknown"
	}
	output, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		// 键不存在时 defaults 返回非零退出码，代表系统处于浅色外观。
		return "light"
	}
	if strings.EqualFold(strings.TrimSpace(string(output)), "dark") {
		return "dark"
	}
	return "light"
}
