package test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/flyerxp/lib/v2/errorsL"
)

// TestNew 验证 New 创建带堆栈的基础错误
func TestNew(t *testing.T) {
	msg := "track test base"
	err := errorsL.New(msg)

	// 1. 错误消息正确
	if err.Error() != msg {
		t.Errorf("New() 错误消息不匹配: got=%q, want=%q", err.Error(), msg)
	}

	// 2. 堆栈输出包含当前测试函数
	formatted := fmt.Sprintf("%+v", err)
	if !strings.Contains(formatted, t.Name()) {
		t.Errorf("New() 未包含调用堆栈，输出:\n%s", formatted)
	}
}

// TestErrorf 验证格式化错误创建 + %w 错误链能力
func TestErrorf(t *testing.T) {
	rootErr := errors.New("root original error")
	err := errorsL.Errorf("wrap info %s: %w", "param", rootErr)

	expectedMsg := "wrap info param: root original error"
	if err.Error() != expectedMsg {
		t.Errorf("Errorf() 消息不匹配: got=%q, want=%q", err.Error(), expectedMsg)
	}

	// 验证标准库错误链穿透
	if !errors.Is(err, rootErr) {
		t.Error("Errorf() 不支持 %w 错误链穿透")
	}
}

// TestWrap 验证错误包装 + 堆栈附加
func TestWrap(t *testing.T) {
	rootErr := errors.New("db connect failed")
	wrapped := errorsL.Wrap(rootErr, "init user module")

	expected := "init user module: db connect failed"
	if wrapped.Error() != expected {
		t.Errorf("Wrap() 消息不匹配: got=%q, want=%q", wrapped.Error(), expected)
	}

	// 错误链可追溯到根错误
	if !errors.Is(wrapped, rootErr) {
		t.Error("Wrap() 错误链断裂，无法追溯根错误")
	}

	// 堆栈记录在 Wrap 调用处
	formatted := fmt.Sprintf("%+v", wrapped)
	if !strings.Contains(formatted, t.Name()) {
		t.Errorf("Wrap() 堆栈未记录当前调用位置，输出:\n%s", formatted)
	}
}

// TestWrapf 验证格式化包装错误
func TestWrapf(t *testing.T) {
	rootErr := errors.New("record not found")
	wrapped := errorsL.Wrapf(rootErr, "query user id=%d failed", 10086)

	expected := "query user id=10086 failed: record not found"
	if wrapped.Error() != expected {
		t.Errorf("Wrapf() 消息不匹配: got=%q, want=%q", wrapped.Error(), expected)
	}

	if !errors.Is(wrapped, rootErr) {
		t.Error("Wrapf() 错误链断裂")
	}
}

// TestWithStack 验证仅附加堆栈、不修改错误消息
func TestWithStack(t *testing.T) {
	plainErr := errors.New("plain std error")
	withStack := errorsL.WithStack(plainErr)

	// 错误文本完全不变
	if withStack.Error() != plainErr.Error() {
		t.Errorf("WithStack() 修改了错误消息: got=%q, want=%q", withStack.Error(), plainErr.Error())
	}

	// 附加了调用堆栈
	formatted := fmt.Sprintf("%+v", withStack)
	if !strings.Contains(formatted, t.Name()) {
		t.Errorf("WithStack() 未附加堆栈，输出:\n%s", formatted)
	}
}

// TestErrorChain 验证多层包装下的 Unwrap 与错误穿透
func TestErrorChain(t *testing.T) {
	root := errors.New("root cause")
	layer1 := errorsL.Wrap(root, "service layer")
	layer2 := errorsL.Wrap(layer1, "handler layer")

	// 逐层 Unwrap 可到达根错误
	unwrap1 := errors.Unwrap(layer2)
	if unwrap1.Error() != "service layer: root cause" {
		t.Errorf("第一层Unwrap结果错误: got=%q", unwrap1.Error())
	}

	unwrapRoot := errors.Unwrap(errors.Unwrap(layer2))
	if unwrapRoot != root {
		t.Error("多层Unwrap无法到达根错误")
	}

	// errors.Is 穿透所有包装层
	if !errors.Is(layer2, root) {
		t.Error("errors.Is 无法穿透多层包装")
	}
}

// TestNilInput 验证 nil 入参的边界处理
func TestNilInput(t *testing.T) {
	if errorsL.WithStack(nil) != nil {
		t.Error("WithStack(nil) 应返回 nil")
	}
	if errorsL.Wrap(nil, "message") != nil {
		t.Error("Wrap(nil) 应返回 nil")
	}
	if errorsL.Wrapf(nil, "format %s", "arg") != nil {
		t.Error("Wrapf(nil) 应返回 nil")
	}
}

// TestStackTraceFormat 验证堆栈输出格式（函数名、文件行号）
func TestStackTraceFormat(t *testing.T) {
	err := errorsL.New("stack format check")
	formatted := fmt.Sprintf("%+v", err)

	// 验证包含文件名+行号的典型格式
	if !strings.Contains(formatted, ".go:") {
		t.Error("堆栈输出缺少 文件:行号 信息")
	}

	// 验证包含当前函数名
	pc, _, _, _ := runtime.Caller(0)
	currentFunc := runtime.FuncForPC(pc).Name()
	if !strings.Contains(formatted, currentFunc) {
		t.Errorf("堆栈未包含当前函数 %s，输出:\n%s", currentFunc, formatted)
	}
}
