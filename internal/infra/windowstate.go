package infra

// WindowPlacement 是窗口位置与尺寸的持久化结构，由 Go 侧维护。
// 它写在同一个 settings.json 的 window 字段里，不另建配置文件。
type WindowPlacement struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximised bool `json:"maximised"`
}

// LoadWindowPlacement 读取上次保存的窗口位置；不存在时 ok 为 false。
func (s *SettingsStore) LoadWindowPlacement() (WindowPlacement, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := s.loadFile()
	if err != nil || file.Window == nil {
		return WindowPlacement{}, false
	}
	return *file.Window, true
}

// SaveWindowPlacement 写入窗口位置，并保留已有的主题/语言设置。
func (s *SettingsStore) SaveWindowPlacement(placement WindowPlacement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := s.loadFile()
	if err != nil {
		return err
	}
	file.Window = &placement
	return s.write(file)
}
