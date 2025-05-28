package u

import (
	"fmt"
	"runtime"
	"strings"
)

type CallStack struct {
	frames []runtime.Frame
}

func (s CallStack) StringSingeLine() string {
	output := new(strings.Builder)
	prefix := ""
	for _, frame := range s.frames {
		output.WriteString(fmt.Sprintf("%s%s:%d %s", prefix, frame.File, frame.Line, frame.Function))
		if prefix == "" {
			prefix = " <- "
		}
	}
	return output.String()
}

func (s CallStack) StringMultiline() string {
	output := new(strings.Builder)
	prefix := ""
	for _, frame := range s.frames {
		output.WriteString(fmt.Sprintf("%s%s:%d %s\n", prefix, frame.File, frame.Line, frame.Function))
		if prefix == "" {
			prefix = " at "
		}
	}
	return output.String()
}

func CurrentCallStack(skip int) CallStack {
	// Add 2 to skip Callers itself and printStack
	pc := make([]uintptr, 20)
	n := runtime.Callers(skip+2, pc) // skip+2 to skip internal frames
	frames := runtime.CallersFrames(pc[:n])
	cs := CallStack{}
	for {
		frame, more := frames.Next()
		cs.frames = append(cs.frames, frame)
		if !more {
			break
		}
	}
	return cs
}
