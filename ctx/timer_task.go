package ctx

import (
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
	"time"
)

const timeTaskTag = "TimeTask"

type TimerTask struct {
	closer chan bool
}

func (instance *TimerTask) StartTimer(interval time.Duration, actionOnTimer func()) {
	instance.closer = make(chan bool)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-instance.closer:
				ticker.Stop()
				return
			case <-ticker.C:
				if err := nopanic.Run(actionOnTimer); err != nil {
					logger.Error(timeTaskTag, err.Error())
				}
			}
		}
	}()
}

func (instance *TimerTask) StopTimer() {
	if instance.closer != nil {
		instance.closer <- true
	}
}
