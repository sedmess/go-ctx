package nopanic

import (
	"errors"
	"fmt"

	"github.com/sedmess/go-ctx/u"
)

type PanicWrapperError interface {
	Error() string
	Stack() u.CallStack
	Reason() string
	ToReasonError() error
}

type panicWrapperError struct {
	reason string
	stack  u.CallStack
}

func (p *panicWrapperError) Error() string {
	return "panic: " + p.reason + ": " + p.stack.StringSingeLine()
}

func (p *panicWrapperError) Stack() u.CallStack {
	return p.stack
}

func (p *panicWrapperError) Reason() string {
	return p.reason
}

func (p *panicWrapperError) ToReasonError() error {
	return errors.New(p.Reason())
}

// IsPanicWrapperError reports whether err directly or transitively wraps a panic captured by
// Run, RunE, or RunResult.
func IsPanicWrapperError(err error) bool {
	var panicErr PanicWrapperError
	return errors.As(err, &panicErr)
}

func wrapPanicReason(reason any) PanicWrapperError {
	return &panicWrapperError{reason: fmt.Sprint(reason), stack: u.CurrentCallStack(3)}
}

func Run(fn func()) (err PanicWrapperError) {
	defer func() {
		if r := recover(); r != nil {
			err = wrapPanicReason(r)
		}
	}()
	fn()
	return
}

func RunE(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wrapPanicReason(r)
		}
	}()
	err = fn()
	return
}

func RunResult[T any](fn func() (T, error)) (res T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wrapPanicReason(r)
		}
	}()
	res, err = fn()
	return
}
