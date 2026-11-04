package contract

// 事件名统一使用「域.动作」，前端只在 bridge 层订阅。
const (
	EventClock  = "app.clock"
	EventStatus = "app.status"
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
