package infra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rouarn/wails3-template/internal/contract"
)

func newTestStore(t *testing.T) (*SettingsStore, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	return NewSettingsStore(dir, path), path
}

func TestSettingsLoadMissingUsesDefaults(t *testing.T) {
	store, _ := newTestStore(t)
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("文件缺失不应返回错误: %v", err)
	}
	defaults := contract.DefaultSettings()
	if settings.Theme != defaults.Theme || settings.Language != defaults.Language {
		t.Fatalf("文件缺失应回退默认值，得到 %+v", settings)
	}
}

func TestSettingsLoadCorruptUsesDefaults(t *testing.T) {
	store, path := newTestStore(t)
	if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
		t.Fatalf("准备损坏文件失败: %v", err)
	}
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("损坏文件应回退默认值: %v", err)
	}
	defaults := contract.DefaultSettings()
	if settings.Theme != defaults.Theme || settings.Language != defaults.Language {
		t.Fatalf("损坏文件应回退默认值，得到 %+v", settings)
	}
}

func TestSettingsSaveLoadRoundTrip(t *testing.T) {
	store, path := newTestStore(t)
	want := contract.Settings{Theme: contract.ThemeDark, Language: contract.LanguageEnUS, UpdatedAt: "2026-11-04T12:00:00Z"}
	if err := store.Save(want); err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got != want {
		t.Fatalf("往返结果 = %+v，期望 %+v", got, want)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("原子写入不应残留 .tmp 文件")
	}
}

func TestWindowPlacementRoundTrip(t *testing.T) {
	store, _ := newTestStore(t)
	want := WindowPlacement{X: 120, Y: 80, Width: 1180, Height: 760, Maximised: true}
	if err := store.SaveWindowPlacement(want); err != nil {
		t.Fatalf("保存窗口位置失败: %v", err)
	}
	got, ok := store.LoadWindowPlacement()
	if !ok {
		t.Fatal("应能读到窗口位置")
	}
	if got != want {
		t.Fatalf("窗口位置 = %+v，期望 %+v", got, want)
	}
}

func TestSaveSettingsPreservesWindowPlacement(t *testing.T) {
	store, _ := newTestStore(t)
	placement := WindowPlacement{X: 10, Y: 20, Width: 1180, Height: 760}
	if err := store.SaveWindowPlacement(placement); err != nil {
		t.Fatalf("保存窗口位置失败: %v", err)
	}
	if err := store.Save(contract.Settings{Theme: contract.ThemeLight, Language: contract.LanguageZhCN, UpdatedAt: "2026-11-04T12:00:00Z"}); err != nil {
		t.Fatalf("保存设置失败: %v", err)
	}
	got, ok := store.LoadWindowPlacement()
	if !ok || got != placement {
		t.Fatalf("保存设置后窗口位置应保留，得到 %+v ok=%v", got, ok)
	}
}
