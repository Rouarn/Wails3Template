package app

import (
	"embed"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"github.com/rouarn/wails3-template/internal/contract"
	"github.com/rouarn/wails3-template/internal/infra"
	"github.com/rouarn/wails3-template/internal/platform"
	"github.com/rouarn/wails3-template/internal/service"
)

// 托盘图标复用品牌四芒星：由 build/appicon.png 缩放到 64px，构建期不需要图像处理依赖。
//
//go:embed assets/trayicon.png
var trayIcon []byte

const (
	// AppName 同时用于窗口标题、配置目录名与日志标识。
	AppName = "Wails3Template"
	// AppVersion 与 build/config.yml 的 info.version 保持一致。
	AppVersion = "0.1.0"
	// WailsVersion 与 go.mod 中固定的 Wails 版本保持一致。
	WailsVersion = "v3.1.0"

	mainWindowName = "main"
)

// Run 装配依赖并进入应用主循环，返回时进程退出。
func Run(assets embed.FS) error {
	dirs, err := platform.NewDirs(AppName)
	if err != nil {
		return err
	}
	logger, closeLogger, err := infra.NewLogger(dirs.LogDir)
	if err != nil {
		return err
	}
	defer closeLogger()

	settingsStore := infra.NewSettingsStore(dirs.ConfigDir, dirs.SettingsFile)

	var instance *application.App
	instance = application.New(application.Options{
		Name:        AppName,
		Description: "基于 Wails v3 与 Vite Plus 的桌面应用模板",
		Logger:      logger,
		LogLevel:    slog.LevelInfo,
		// Go error → {code,message,details}，前端 bridge 层统一消费。
		MarshalError: contract.MarshalError,
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: securityHeaders,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.rouarn.wails3-template",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				logger.Info("检测到第二个实例，已聚焦主窗口", "args", data.Args)
				if window, ok := instance.Window.GetByName(mainWindowName); ok {
					window.Show()
					window.Focus()
				}
			},
		},
		Mac: application.MacOptions{
			// 关窗只收进托盘；只有托盘「退出」才结束进程。
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		ErrorHandler: func(err error) {
			logger.Error("应用运行错误", "error", err)
		},
	})

	// 先以隐藏+居中创建；屏幕信息要等 ApplicationStarted 后才由运行时缓存，
	// 因此位置恢复放到启动回调里，避免用空的屏幕列表误判「窗口已在屏幕外」。
	mainWindow := instance.Window.NewWithOptions(mainWindowOptions())

	// 服务在 app 创建之后再注册，这样窗口与服务共享同一个实例，无需 DI 框架。
	systemService := service.NewSystemService(instance, dirs, AppVersion, WailsVersion)
	settingsService := service.NewSettingsService(instance, settingsStore)
	windowService := service.NewWindowService(mainWindow)
	capabilitiesService := service.NewCapabilitiesService(instance, mainWindow, systemService, settingsStore, notifications.New())
	instance.RegisterService(application.NewService(systemService))
	instance.RegisterService(application.NewService(settingsService))
	instance.RegisterService(application.NewService(windowService))
	instance.RegisterService(application.NewService(capabilitiesService))

	var quitting atomic.Bool
	// 窗口位置只在移动/缩放时采集，退出阶段用缓存落盘，避免读到已销毁窗口的 0 值。
	tracker := newPlacementTracker(settingsStore, logger)
	mainWindow.OnWindowEvent(events.Common.WindowDidMove, func(*application.WindowEvent) {
		tracker.observe(mainWindow)
	})
	mainWindow.OnWindowEvent(events.Common.WindowDidResize, func(*application.WindowEvent) {
		tracker.observe(mainWindow)
	})
	instance.OnShutdown(tracker.persist)
	mainWindow.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if quitting.Load() {
			return
		}
		event.Cancel()
		tracker.persist()
		mainWindow.Hide()
	})
	setupTray(instance, mainWindow, &quitting)

	instance.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		go func() {
			applyWindowPlacement(instance, mainWindow, settingsStore, logger)
			tracker.observe(mainWindow)
			mainWindow.Show()
			mainWindow.Focus()
		}()
	})

	startedAt := time.Now()
	go emitClock(instance, startedAt)

	logger.Info("应用启动", "version", AppVersion, "wails", WailsVersion, "config", dirs.ConfigDir)
	return instance.Run()
}

// 窗口默认尺寸与最小尺寸；恢复历史位置时也按这套下限校验，过小视为无效记录。
const (
	windowDefaultWidth  = 1180
	windowDefaultHeight = 760
	windowMinWidth      = 900
	windowMinHeight     = 620
)

// mainWindowOptions 返回首次启动的默认窗口选项：隐藏创建、居中、固定默认尺寸。
func mainWindowOptions() application.WebviewWindowOptions {
	return application.WebviewWindowOptions{
		Name:             mainWindowName,
		Title:            "Wails v3 桌面模板",
		Width:            windowDefaultWidth,
		Height:           windowDefaultHeight,
		MinWidth:         windowMinWidth,
		MinHeight:        windowMinHeight,
		InitialPosition:  application.WindowCentered,
		Hidden:           true,
		URL:              "/",
		BackgroundColour: application.NewRGB(16, 18, 23),
	}
}

// applyWindowPlacement 在启动回调里恢复上次的位置与尺寸。
// 屏幕列表在 ApplicationStarted 阶段才可用，这里做有限重试；仍为空或位置不可见时保持居中。
func applyWindowPlacement(instance *application.App, window application.Window, store *infra.SettingsStore, logger *slog.Logger) {
	placement, ok := store.LoadWindowPlacement()
	if !ok {
		return
	}
	var screens []*application.Screen
	for attempt := 0; attempt < 20 && len(screens) == 0; attempt++ {
		screens = instance.Screen.GetAll()
		if len(screens) == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !placementVisible(placement, screens) {
		logger.Info("窗口位置不可见，回落到居中", "x", placement.X, "y", placement.Y)
		return
	}
	window.SetSize(placement.Width, placement.Height)
	window.SetPosition(placement.X, placement.Y)
	if placement.Maximised {
		window.Maximise()
	}
}

// placementVisible 用窗口中心点判断是否落在任一显示器工作区内，避免拔掉外接屏后窗口“消失”。
func placementVisible(placement infra.WindowPlacement, screens []*application.Screen) bool {
	if placement.Width < windowMinWidth || placement.Height < windowMinHeight {
		return false
	}
	centerX := placement.X + placement.Width/2
	centerY := placement.Y + placement.Height/2
	for _, screen := range screens {
		area := screen.WorkArea
		if centerX >= area.X && centerX < area.X+area.Width && centerY >= area.Y && centerY < area.Y+area.Height {
			return true
		}
	}
	return false
}

// setupTray 注册托盘与中文菜单：显示主窗口 / 设置 / 退出。
func setupTray(instance *application.App, window application.Window, quitting *atomic.Bool) {
	tray := instance.SystemTray.New()
	tray.SetTooltip(AppName)
	if len(trayIcon) > 0 {
		tray.SetIcon(trayIcon)
	}
	menu := application.NewMenu()
	menu.Add("显示主窗口").OnClick(func(*application.Context) {
		window.Show()
		window.Focus()
	})
	menu.Add("设置").OnClick(func(*application.Context) {
		window.Show()
		window.Focus()
		instance.Event.Emit(contract.EventNavigate, contract.NavigateEvent{Path: "/settings"})
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(*application.Context) {
		quitting.Store(true)
		instance.Quit()
	})
	tray.SetMenu(menu)
}

// emitClock 每秒推送一次时间与运行时长，演示后端 → 前端的事件通道。
func emitClock(instance *application.App, startedAt time.Time) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		instance.Event.Emit(contract.EventClock, contract.ClockEvent{
			Now:           now.Format(time.RFC3339),
			UnixMillis:    now.UnixMilli(),
			UptimeSeconds: int64(now.Sub(startedAt).Seconds()),
		})
	}
}
