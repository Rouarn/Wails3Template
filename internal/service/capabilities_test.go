package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rouarn/wails3-template/internal/contract"
)

func TestRunDemoTaskSuccess(t *testing.T) {
	var events []contract.TaskProgressEvent
	result, err := runDemoTask(context.Background(), 3, 50, func(event contract.TaskProgressEvent) {
		events = append(events, event)
	}, func(context.Context, time.Duration) bool { return true })
	if err != nil {
		t.Fatalf("正常任务不应报错: %v", err)
	}
	if !result.Completed || result.Steps != 3 {
		t.Fatalf("结果 = %+v，期望 Steps=3 Completed=true", result)
	}
	if len(events) != 3 {
		t.Fatalf("进度事件数 = %d，期望 3", len(events))
	}
	for index, event := range events {
		step := index + 1
		if event.Step != step || event.Total != 3 || event.Percent != step*100/3 || event.State != "running" {
			t.Fatalf("第 %d 个事件 = %+v", step, event)
		}
	}
}

func TestRunDemoTaskCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var events []contract.TaskProgressEvent
	_, err := runDemoTask(ctx, 3, 50, func(event contract.TaskProgressEvent) {
		events = append(events, event)
	}, func(context.Context, time.Duration) bool { return false })
	if err == nil {
		t.Fatal("取消后应返回错误")
	}
	var appErr *contract.AppError
	if !errors.As(err, &appErr) || appErr.Code != contract.CodeCancelled {
		t.Fatalf("错误 = %v，期望 code=%s", err, contract.CodeCancelled)
	}
	if len(events) != 1 || events[0].State != "cancelled" || events[0].Step != 0 {
		t.Fatalf("取消事件 = %+v", events)
	}
}

func TestRunDemoTaskRejectsInvalidArguments(t *testing.T) {
	cases := []struct {
		name    string
		steps   int
		delayMs int
	}{
		{"steps 过小", 0, 50},
		{"steps 过大", 21, 50},
		{"delay 过小", 3, 49},
		{"delay 过大", 3, 2001},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := runDemoTask(context.Background(), testCase.steps, testCase.delayMs, func(contract.TaskProgressEvent) {}, func(context.Context, time.Duration) bool { return true })
			var appErr *contract.AppError
			if !errors.As(err, &appErr) || appErr.Code != contract.CodeInvalidArgument {
				t.Fatalf("错误 = %v，期望 code=%s", err, contract.CodeInvalidArgument)
			}
		})
	}
}

func TestWaitOrCancel(t *testing.T) {
	if ok := waitOrCancel(context.Background(), 10*time.Millisecond); !ok {
		t.Fatal("未取消时应等待完成并返回 true")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if ok := waitOrCancel(ctx, time.Hour); ok {
		t.Fatal("已取消时应返回 false")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("取消应立即返回，实际耗时 %v", elapsed)
	}
}
