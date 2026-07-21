package ctx

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type manualTicker struct {
	ticks    chan time.Time
	stopped  chan struct{}
	stopOnce sync.Once
}

func newManualTicker() *manualTicker {
	return &manualTicker{ticks: make(chan time.Time), stopped: make(chan struct{})}
}

func (ticker *manualTicker) handle() timerTicker {
	return timerTicker{
		ticks: ticker.ticks,
		stop:  func() { ticker.stopOnce.Do(func() { close(ticker.stopped) }) },
	}
}

func TestTimerTaskStopReplacementAndRestart(t *testing.T) {
	created := make(chan *manualTicker, 3)
	originalFactory := timerTickerFactory
	timerTickerFactory = func(time.Duration) timerTicker {
		ticker := newManualTicker()
		created <- ticker
		return ticker.handle()
	}
	defer func() { timerTickerFactory = originalFactory }()
	var timer TimerTask
	timer.StopTimer()

	oldTicks := atomic.Int32{}
	timer.StartTimer(time.Hour, func() { oldTicks.Add(1) })
	first := <-created
	first.ticks <- time.Time{}
	waitForValue(t, &oldTicks, 1, "first timer did not tick")

	replacementTicks := atomic.Int32{}
	timer.StartTimer(time.Hour, func() { replacementTicks.Add(1) })
	select {
	case <-first.stopped:
	default:
		t.Fatal("replacement did not stop and join the first ticker")
	}
	second := <-created
	second.ticks <- time.Time{}
	waitForValue(t, &replacementTicks, 1, "replacement timer did not tick")
	timer.StopTimer()
	<-second.stopped
	timer.StopTimer()

	restartedTicks := atomic.Int32{}
	timer.StartTimer(time.Hour, func() { restartedTicks.Add(1) })
	third := <-created
	third.ticks <- time.Time{}
	waitForValue(t, &restartedTicks, 1, "restarted timer did not tick")
	timer.StopTimer()
	<-third.stopped
}

func TestTimerTaskStopWaitsForInflightAction(t *testing.T) {
	ticker := newManualTicker()
	originalFactory := timerTickerFactory
	timerTickerFactory = func(time.Duration) timerTicker { return ticker.handle() }
	defer func() { timerTickerFactory = originalFactory }()
	var timer TimerTask
	entered := make(chan struct{})
	release := make(chan struct{})
	timer.StartTimer(time.Hour, func() {
		close(entered)
		<-release
	})
	ticker.ticks <- time.Time{}
	<-entered

	timer.opMu.Lock()
	generation := timer.generation
	timer.opMu.Unlock()
	stopped := make(chan struct{})
	go func() {
		timer.StopTimer()
		close(stopped)
	}()
	<-generation.stop
	select {
	case <-stopped:
		t.Fatal("StopTimer returned before the action completed")
	default:
	}
	close(release)
	waitForTest(t, stopped, "StopTimer did not join the worker")
}

func waitForValue(t *testing.T, value *atomic.Int32, want int32, message string) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for value.Load() != want {
		select {
		case <-deadline.C:
			t.Fatal(message)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
