package ctx

import (
	"sync"
	"time"

	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
)

const timeTaskTag = "TimeTask"

type timerGeneration struct {
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type timerTicker struct {
	ticks <-chan time.Time
	stop  func()
}

func createTimerTicker(interval time.Duration) timerTicker {
	ticker := time.NewTicker(interval)
	return timerTicker{ticks: ticker.C, stop: ticker.Stop}
}

var timerTickerFactory = createTimerTicker

func (generation *timerGeneration) stopAndJoin() {
	if generation == nil {
		return
	}
	generation.stopOnce.Do(func() { close(generation.stop) })
	<-generation.done
}

type TimerTask struct {
	opMu       sync.Mutex
	generation *timerGeneration
}

// StartTimer replaces and joins an existing timer generation before starting the next one.
func (instance *TimerTask) StartTimer(interval time.Duration, actionOnTimer func()) {
	instance.opMu.Lock()
	defer instance.opMu.Unlock()
	instance.generation.stopAndJoin()

	ticker := timerTickerFactory(interval)
	generation := &timerGeneration{stop: make(chan struct{}), done: make(chan struct{})}
	instance.generation = generation
	go func() {
		defer close(generation.done)
		defer ticker.stop()
		for {
			select {
			case <-generation.stop:
				return
			default:
			}
			select {
			case <-generation.stop:
				return
			case <-ticker.ticks:
				if err := nopanic.Run(actionOnTimer); err != nil {
					logger.Error(timeTaskTag, err.Error())
				}
			}
		}
	}()
}

// StopTimer is idempotent and waits for an in-flight action and worker exit.
func (instance *TimerTask) StopTimer() {
	instance.opMu.Lock()
	defer instance.opMu.Unlock()
	instance.generation.stopAndJoin()
	instance.generation = nil
}
