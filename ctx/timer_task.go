package ctx

import "time"

type TimerTask struct {
	closer chan bool
}

func (instance TimerTask) StartTimer(interval time.Duration, actionOnTimer func()) {
	instance.closer = make(chan bool)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-instance.closer:
				ticker.Stop()
				return
			case <-ticker.C:
				actionOnTimer()
			}
		}
	}()
}

func (instance TimerTask) StopTimer() {
	if instance.closer != nil {
		instance.closer <- true
	}
}
