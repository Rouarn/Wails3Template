package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rouarn/wails3-template/internal/contract"
)

// SettingsStore 负责设置文件的读取与原子写入。
// 文件缺失视为首次启动，返回默认设置。
type SettingsStore struct {
	mu   sync.Mutex
	dir  string
	path string
}

// NewSettingsStore 创建 store，dir 通常为 platform.Dirs.ConfigDir。
func NewSettingsStore(dir string, file string) *SettingsStore {
	return &SettingsStore{dir: dir, path: file}
}

// Load 读取设置；文件不存在或损坏时回退默认值。
func (s *SettingsStore) Load() (contract.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return contract.DefaultSettings(), nil
		}
		return contract.Settings{}, fmt.Errorf("读取设置失败: %w", err)
	}
	var settings contract.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return contract.DefaultSettings(), nil
	}
	return settings, nil
}

// Save 先写临时文件再 rename，避免写入中断留下半截 JSON。
func (s *SettingsStore) Save(settings contract.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("创建设置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化设置失败: %w", err)
	}
	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("写入临时设置文件失败: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("替换设置文件失败: %w", err)
	}
	return nil
}

// Path 返回设置文件的绝对路径，供运行信息页展示。
func (s *SettingsStore) Path() string {
	return filepath.Clean(s.path)
}
