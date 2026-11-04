package contract

import "testing"

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()
	if settings.Theme != ThemeSystem {
		t.Fatalf("默认主题 = %q，期望 %q", settings.Theme, ThemeSystem)
	}
	if settings.Language != LanguageZhCN {
		t.Fatalf("默认语言 = %q，期望 %q", settings.Language, LanguageZhCN)
	}
	// 默认值不含 updatedAt：它只在保存/重置时由 SettingsService 写入。
	if settings.UpdatedAt != "" {
		t.Fatalf("默认设置不应带 updatedAt，实际 %q", settings.UpdatedAt)
	}
}

func TestIsSupportedTheme(t *testing.T) {
	for _, theme := range []string{ThemeSystem, ThemeLight, ThemeDark} {
		if !IsSupportedTheme(theme) {
			t.Fatalf("%q 应被支持", theme)
		}
	}
	if IsSupportedTheme("neon") {
		t.Fatal("未知主题不应被支持")
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	for _, language := range []string{LanguageZhCN, LanguageEnUS} {
		if !IsSupportedLanguage(language) {
			t.Fatalf("%q 应被支持", language)
		}
	}
	if IsSupportedLanguage("ja-JP") {
		t.Fatal("未知语言不应被支持")
	}
}
