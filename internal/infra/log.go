package infra

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	logFileName  = "app.log"
	logMaxBytes  = 5 * 1024 * 1024
	logKeepCount = 3
)

// NewLogger 返回同时写 stdout 与日志文件的 slog.Logger。
// 关闭函数由调用方 defer 执行，确保刷新缓冲区。
func NewLogger(logDir string) (*slog.Logger, func(), error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	logPath := filepath.Join(logDir, logFileName)
	rotateIfNeeded(logPath)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	writer := io.MultiWriter(os.Stdout, file)
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Value.Kind() == slog.KindString {
				attr.Value = slog.StringValue(Redact(attr.Value.String()))
			}
			return attr
		},
	})
	return slog.New(handler), func() { _ = file.Close() }, nil
}

// Redact 去掉日志里的用户主目录与常见凭据形态。
func Redact(value string) string {
	if value == "" {
		return value
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		value = strings.ReplaceAll(value, home, "~")
	}
	if index := strings.Index(value, "token="); index >= 0 {
		value = value[:index] + "token=<已脱敏>"
	}
	return value
}

// rotateIfNeeded 采用 app.log → app.log.1 → app.log.2 的简单轮转，
// 只保留 logKeepCount 份历史，避免日志无限增长。
func rotateIfNeeded(logPath string) {
	info, err := os.Stat(logPath)
	if err != nil || info.Size() < logMaxBytes {
		return
	}
	for index := logKeepCount - 1; index >= 1; index-- {
		oldPath := fmt.Sprintf("%s.%d", logPath, index)
		newPath := fmt.Sprintf("%s.%d", logPath, index+1)
		_ = os.Rename(oldPath, newPath)
	}
	_ = os.Rename(logPath, logPath+".1")
}
