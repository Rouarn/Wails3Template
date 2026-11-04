package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestMarshalErrorStructured(t *testing.T) {
	original := NewAppError(CodeNotFound, "文件不存在")
	data := MarshalError(original)
	if data == nil {
		t.Fatal("结构化错误不应返回 nil")
	}
	var got AppError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if got.Code != CodeNotFound || got.Message != "文件不存在" {
		t.Fatalf("错误字段不匹配: %+v", got)
	}
}

func TestMarshalErrorWrapped(t *testing.T) {
	original := NewAppError(CodePermissionDenied, "没有权限")
	data := MarshalError(fmt.Errorf("调用失败: %w", original))
	if data == nil {
		t.Fatal("包装后的结构化错误不应返回 nil")
	}
	var got AppError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if got.Code != CodePermissionDenied {
		t.Fatalf("errors.As 未还原结构化错误，得到 %+v", got)
	}
}

func TestMarshalErrorGeneric(t *testing.T) {
	data := MarshalError(errors.New("磁盘写入失败"))
	if data == nil {
		t.Fatal("未识别错误应包成 INTERNAL，而不是 nil")
	}
	var got AppError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if got.Code != CodeInternal || got.Message != "磁盘写入失败" {
		t.Fatalf("未识别错误应回退 INTERNAL，得到 %+v", got)
	}
}

func TestMarshalErrorNil(t *testing.T) {
	if data := MarshalError(nil); data != nil {
		t.Fatalf("nil 错误应返回 nil，得到 %s", data)
	}
}

func TestNewInvalidArgument(t *testing.T) {
	err := NewInvalidArgument("steps 超出范围", "允许 1–20")
	if err.Code != CodeInvalidArgument {
		t.Fatalf("错误码 = %q，期望 %q", err.Code, CodeInvalidArgument)
	}
	if err.Details != "允许 1–20" {
		t.Fatalf("details = %q", err.Details)
	}
	if err.Error() != CodeInvalidArgument+": steps 超出范围" {
		t.Fatalf("Error() = %q", err.Error())
	}
}
