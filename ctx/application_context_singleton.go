package ctx

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var globalLock sync.Mutex
var ctx *appContext

type Application interface {
	Stop() Application
	Join()
}

type application struct {
	stopCh chan bool
	mu     sync.RWMutex
}

func (a *application) Stop() Application {
	a.stopCh <- true
	return a
}

func (a *application) Join() {
	a.mu.RLock()
	a.mu.RUnlock()
}

// Deprecated: use ctx.StartApplicationBlocking
func StartContextualizedApplication(servicePackages ...[]any) {
	StartApplication(servicePackages...).Join()
}

func StartApplication(servicePackages ...[]any) Application {
	ctxInstance := func() *appContext {
		globalLock.Lock()
		defer globalLock.Unlock()

		ctxInstance := newApplicationContext()
		for _, services := range servicePackages {
			for _, service := range services {
				ctxInstance.register(service)
			}
		}

		ctxInstance.start()

		ctx = ctxInstance

		return ctxInstance
	}()

	ctxInstanceStoppedCh := make(chan bool)
	go ctxInstance.eventLoop(ctxInstanceStoppedCh)

	app := application{
		stopCh: make(chan bool),
		mu:     sync.RWMutex{},
	}

	go func() {
		defer func() {
			globalLock.Lock()
			defer globalLock.Unlock()

			ctx = nil
		}()

		osSignalCh := make(chan os.Signal)
		signal.Notify(osSignalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		select {
		case <-osSignalCh:
			break
		case <-app.stopCh:
			break
		}

		ctxInstance.eventBus <- event{kind: eStop}
		<-ctxInstanceStoppedCh
		app.mu.Unlock()
	}()

	app.mu.Lock()

	return &app
}

func GetService(serviceName string) any {
	globalLock.Lock()
	defer globalLock.Unlock()

	if ctx != nil {
		return ctx.GetService(serviceName)
	} else {
		panic("no active context")
	}
}

func sendEvent(e event) {
	globalLock.Lock()
	defer globalLock.Unlock()

	if ctx != nil {
		ctx.sendEvent(e)
	} else {
		panic("no active context")
	}
}
