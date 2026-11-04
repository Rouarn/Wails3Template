package app

import (
	"log/slog"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/rouarn/wails3-template/internal/infra"
)

// placementTracker 缓存最近一次有效的窗口位置。
// 退出阶段窗口可能已被销毁（Size/Position 返回 0），直接读取会把 0×0 写进设置文件，
// 因此只在移动/缩放事件里采集，落盘时只用缓存值。
type placementTracker struct {
	mu     sync.Mutex
	store  *infra.SettingsStore
	logger *slog.Logger
	last   infra.WindowPlacement
	valid  bool
}

func newPlacementTracker(store *infra.SettingsStore, logger *slog.Logger) *placementTracker {
	return &placementTracker{store: store, logger: logger}
}

// observe 在窗口移动/缩放时更新内存缓存，不落盘，避免拖动期间频繁写文件。
func (t *placementTracker) observe(window application.Window) {
	width, height := window.Size()
	x, y := window.Position()
	if width <= 0 || height <= 0 {
		return
	}
	t.mu.Lock()
	t.last = infra.WindowPlacement{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximised: window.IsMaximised(),
	}
	t.valid = true
	t.mu.Unlock()
}

// persist 把缓存的位置写入设置；没有采集到有效值时直接跳过。
func (t *placementTracker) persist() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.valid {
		return
	}
	if err := t.store.SaveWindowPlacement(t.last); err != nil {
		t.logger.Warn("保存窗口位置失败", "error", err)
	}
}
