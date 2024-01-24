package ctx

import (
	"github.com/sedmess/go-ctx/logger"
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
	logger.Init(logger.DEBUG)
	exitCode := m.Run()
	os.Exit(exitCode)
}
