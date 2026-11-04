import type { AppInfo, ClockEvent, Settings, StatusEvent } from '@bindings/github.com/rouarn/wails3-template/internal/contract/models'

// 视图层只从 bridge 取类型，不直接 import bindings，避免生成产物渗透到业务代码。
export type { AppInfo, ClockEvent, Settings, StatusEvent }
