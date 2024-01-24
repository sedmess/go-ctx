package ctx

import "runtime/debug"

func Run(fn func()) {
	go func() {
		defer func() {
			if panicReason := recover(); panicReason != nil {
				sendEvent(event{kind: eUnhandledPanic, payload: panicPayload{reason: panicReason, stack: debug.Stack()}})
			}
		}()
		fn()
	}()
}

func RunWithRecover(fn func()) {
	go func() {
		defer func() {
			if panicReason := recover(); panicReason != nil {
				sendEvent(event{kind: eSuppressedPanic, payload: panicPayload{reason: panicReason, stack: debug.Stack()}})
			}
		}()
		fn()
	}()
}

func RunFinally(fn func(), finallyFn func()) {
	go func() {
		defer func() {
			finallyFn()
			if panicReason := recover(); panicReason != nil {
				sendEvent(event{kind: eUnhandledPanic, payload: panicPayload{reason: panicReason, stack: debug.Stack()}})
			}
		}()
		fn()
	}()
}
