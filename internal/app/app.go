package app

import (
	"embed"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"github.com/rouarn/wails3-template/internal/contract"
	"github.com/rouarn/wails3-template/internal/infra"
	"github.com/rouarn/wails3-template/internal/platform"
	"github.com/rouarn/wails3-template/internal/service"
)

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
			// 没有托盘实现时，关闭最后一个窗口即退出，避免进程留在后台无法唤回。
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		ErrorHandler: func(err error) {
			logger.Error("应用运行错误", "error", err)
		},
	})

	mainWindow := instance.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             mainWindowName,
		Title:            "Wails v3 桌面模板",
		Width:            1180,
		Height:           760,
		MinWidth:         900,
		MinHeight:        620,
		URL:              "/",
		BackgroundColour: application.NewRGB(16, 18, 23),
	})

	// 服务在 app 创建之后再注册，这样窗口与服务共享同一个实例，无需 DI 框架。
	systemService := service.NewSystemService(instance, dirs, AppVersion, WailsVersion)
	settingsService := service.NewSettingsService(instance, settingsStore)
	windowService := service.NewWindowService(mainWindow)
	capabilitiesService := service.NewCapabilitiesService(instance, mainWindow, systemService, settingsStore, notifications.New())
	instance.RegisterService(application.NewService(systemService))
	instance.RegisterService(application.NewService(settingsService))
	instance.RegisterService(application.NewService(windowService))
	instance.RegisterService(application.NewService(capabilitiesService))

	startedAt := time.Now()
	go emitClock(instance, startedAt)

	logger.Info("应用启动", "version", AppVersion, "wails", WailsVersion, "config", dirs.ConfigDir)
	return instance.Run()
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
