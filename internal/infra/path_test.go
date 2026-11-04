package infra

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rouarn/wails3-template/internal/contract"
)

func assertInvalidArgument(t *testing.T, err error) {
	t.Helper()
	var appErr *contract.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("期望结构化错误，实际 %v", err)
	}
	if appErr.Code != contract.CodeInvalidArgument {
		t.Fatalf("错误码 = %q，期望 %q", appErr.Code, contract.CodeInvalidArgument)
	}
}

func TestResolveWithinRootRelative(t *testing.T) {
	root := t.TempDir()
	resolved, err := ResolveWithinRoot(root, filepath.Join("nested", "file.txt"))
	if err != nil {
		t.Fatalf("根目录内路径不应失败: %v", err)
	}
	if resolved != filepath.Join(root, "nested", "file.txt") {
		t.Fatalf("解析结果 = %q", resolved)
	}
}

func TestResolveWithinRootRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveWithinRoot(root, filepath.Join("..", "outside.txt")); err == nil {
		t.Fatal(".. 越界应被拒绝")
	} else {
		assertInvalidArgument(t, err)
	}
}

func TestResolveWithinRootRejectsAbsoluteOutside(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if _, err := ResolveWithinRoot(root, filepath.Join(outside, "file.txt")); err == nil {
		t.Fatal("根目录外的绝对路径应被拒绝")
	} else {
		assertInvalidArgument(t, err)
	}
}

func TestResolveWithinRootRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows 创建符号链接需要额外权限")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("准备外部文件失败: %v", err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("创建符号链接失败: %v", err)
	}
	if _, err := ResolveWithinRoot(root, filepath.Join("escape", "secret.txt")); err == nil {
		t.Fatal("符号链接逃逸应被拒绝")
	} else {
		assertInvalidArgument(t, err)
	}
}

func TestResolveWithinRootRejectsEmpty(t *testing.T) {
	if _, err := ResolveWithinRoot("", "file.txt"); err == nil {
		t.Fatal("空 root 应被拒绝")
	}
	if _, err := ResolveWithinRoot(t.TempDir(), ""); err == nil {
		t.Fatal("空 path 应被拒绝")
	}
}
