package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/rouarn/wails3-template/internal/app"
	"github.com/rouarn/wails3-template/internal/contract"
)

// 前端产物通过 go:embed 打进二进制；dev 模式下由 WAILS_VITE_PORT 指定的开发服务器接管。
//
//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// 注册强类型事件，生成器会把事件名与数据类型带进 bindings。
	application.RegisterEvent[contract.ClockEvent](contract.EventClock)
	application.RegisterEvent[contract.StatusEvent](contract.EventStatus)
	application.RegisterEvent[contract.NavigateEvent](contract.EventNavigate)
	application.RegisterEvent[contract.TaskProgressEvent](contract.EventTaskProgress)
}

func main() {
	if err := app.Run(assets); err != nil {
		log.Fatal(err)
	}
}
