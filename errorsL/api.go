package errorsL

import (
	"fmt"
	"runtime"
)

// withStack 携带原始错误与调用堆栈，格式化行为与 pkg/errors 完全兼容
type withStack struct {
	err   error
	stack []uintptr // 调用栈PC指针数组
}

// Error 实现 error 接口，返回原始错误信息
func (w *withStack) Error() string {
	return w.err.Error()
}

// Unwrap 支持标准库 errors.Is / errors.As / errors.Unwrap 错误链能力
func (w *withStack) Unwrap() error {
	return w.err
}

// Format 实现 fmt.Formatter 接口
// - %v / %s：仅输出错误文本
// - %+v：输出错误文本 + 完整调用栈，与 pkg/errors 输出格式完全一致
func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, w.err.Error())
			for _, pc := range w.stack {
				fn := runtime.FuncForPC(pc - 1)
				if fn == nil {
					continue
				}
				file, line := fn.FileLine(pc - 1)
				fmt.Fprintf(s, "\n%s\n\t%s:%d", fn.Name(), file, line)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// New 创建带堆栈的新错误，等价于 pkg/errors 的 errors.New
func New(message string) error {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	return &withStack{
		err:   fmt.Errorf("%s", message),
		stack: pcs[:n],
	}
}

// Errorf 创建带堆栈的格式化错误，等价于 pkg/errors 的 errors.Errorf
// 支持 %w 嵌套错误，兼容标准库错误链
func Errorf(format string, args ...interface{}) error {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	return &withStack{
		err:   fmt.Errorf(format, args...),
		stack: pcs[:n],
	}
}

// WithStack 为已有错误附加当前调用栈，等价于 pkg/errors 的 errors.WithStack
func WithStack(err error) error {
	if err == nil {
		return nil
	}
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	return &withStack{err: err, stack: pcs[:n]}
}

// Wrap 包装错误并附加调用栈，等价于 pkg/errors 的 errors.Wrap
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	return &withStack{
		err:   fmt.Errorf("%s: %w", message, err),
		stack: pcs[:n],
	}
}

// Wrapf 格式化包装错误并附加调用栈，等价于 pkg/errors 的 errors.Wrapf
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	return &withStack{
		err:   fmt.Errorf(format+": %w", append(args, err)...),
		stack: pcs[:n],
	}
}
