package ctx

import (
	"os"
	"testing"
)

type lifecycleNotifier struct {
	ch chan bool
}

func newLifecycleNotifier() *lifecycleNotifier {
	return &lifecycleNotifier{
		ch: make(chan bool, 1),
	}
}

func (l *lifecycleNotifier) AfterStart() {
	go func() { l.ch <- true }()
}

func (l *lifecycleNotifier) BeforeStop() {
	go func() { l.ch <- false }()
}

func TestMain(m *testing.M) {
	_ = os.Setenv("SLOG_HANDLER", "legacy")
	_ = os.Setenv("SLOG_LEVEL", "debug")
	exitCode := m.Run()
	os.Exit(exitCode)
}
