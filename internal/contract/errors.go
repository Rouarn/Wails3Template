package contract

import (
	"encoding/json"
	"errors"
)

// 错误码：前端 bridge 层按 code 分流，不解析英文或中文文案。
const (
	CodeInternal         = "INTERNAL"
	CodeInvalidArgument  = "INVALID_ARGUMENT"
	CodeUnsupported      = "UNSUPPORTED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeNotFound         = "NOT_FOUND"
)

// AppError 是前后端共享的结构化错误。
// 通过 application.Options.MarshalError 序列化为 {code,message,details}。
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

// NewAppError 构造结构化错误。
func NewAppError(code string, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// NewInvalidArgument 构造参数校验错误。
func NewInvalidArgument(message string, details string) *AppError {
	return &AppError{Code: CodeInvalidArgument, Message: message, Details: details}
}

// MarshalError 把任意 error 归一化为 JSON 字节。
// 返回 nil 时 Wails 会退回默认错误处理，因此未识别错误也统一包成 INTERNAL。
func MarshalError(err error) []byte {
	if err == nil {
		return nil
	}
	payload := AppError{Code: CodeInternal, Message: err.Error()}
	var appErr *AppError
	if errors.As(err, &appErr) {
		payload = *appErr
	}
	data, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil
	}
	return data
}
