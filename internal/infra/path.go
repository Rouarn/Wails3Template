package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rouarn/wails3-template/internal/contract"
)

// ResolveWithinRoot 把 p 解析为绝对路径，并保证结果不逃出 root。
// 校验分两步：先用 filepath.Rel 拦住 ".."，再对已存在的路径做符号链接比对。
func ResolveWithinRoot(root string, p string) (string, error) {
	if root == "" || p == "" {
		return "", contract.NewInvalidArgument("路径不能为空", "root 与 path 都是必填项")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("解析根目录失败: %w", err)
	}
	candidate := p
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(absRoot, candidate)
	}
	absPath := filepath.Clean(candidate)

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", contract.NewInvalidArgument("路径不在允许的根目录内", p)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", contract.NewInvalidArgument("路径越界", fmt.Sprintf("%s 不在 %s 内", p, root))
	}

	// 符号链接逃逸：仅当目标已存在时才可解析，未创建的路径由上面的 Rel 校验兜底。
	resolvedRoot, rootErr := filepath.EvalSymlinks(absRoot)
	if rootErr != nil {
		return absPath, nil
	}
	resolvedPath, pathErr := filepath.EvalSymlinks(absPath)
	if pathErr != nil {
		return absPath, nil
	}
	resolvedRel, relErr := filepath.Rel(resolvedRoot, resolvedPath)
	if relErr != nil || resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(filepath.Separator)) {
		return "", contract.NewInvalidArgument("路径经符号链接逃逸", fmt.Sprintf("%s 指向 %s", p, resolvedPath))
	}
	return absPath, nil
}

// EnsureDir 创建目录（0o755），已存在时直接返回。
func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	return nil
}
