package contract

// 事件名统一使用「域.动作」，前端只在 bridge 层订阅。
const (
	EventClock        = "app.clock"
	EventStatus       = "app.status"
	EventNavigate     = "app.navigate"
	EventTaskProgress = "app.task.progress"
)

// ClockEvent 每秒推送一次，用于演示后端 → 前端的时间/运行时长推送。
type ClockEvent struct {
	// Now 是 RFC3339 格式的本地时间。
	Now string `json:"now"`
	// UnixMillis 便于前端直接做时间计算。
	UnixMillis int64 `json:"unixMillis"`
	// UptimeSeconds 是应用启动至今的秒数。
	UptimeSeconds int64 `json:"uptimeSeconds"`
}

// StatusEvent 用于推送一次性状态变更（启动完成、设置已保存等）。
type StatusEvent struct {
	State   string `json:"state"`
	Message string `json:"message"`
	At      string `json:"at"`
}

// NavigateEvent 由托盘菜单或系统入口触发，前端据此切换路由。
type NavigateEvent struct {
	Path string `json:"path"`
}

// TaskProgressEvent 是后端任务 → 前端的进度流；它是「stream 示例」的传输方式。
type TaskProgressEvent struct {
	Step    int    `json:"step"`
	Total   int    `json:"total"`
	Percent int    `json:"percent"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// TaskResult 是 RunDemoTask 的最终返回值。
type TaskResult struct {
	Steps     int  `json:"steps"`
	Completed bool `json:"completed"`
}
