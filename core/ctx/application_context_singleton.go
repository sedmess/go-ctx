package ctx

import (
	"github.com/sedmess/go-ctx/logger"
	"github.com/sedmess/go-ctx/u"
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

// Deprecated: use ctx.CreateContextualizedApplication with .Join() call
func StartContextualizedApplication(servicePackages ...[]any) {
	pkgs := make([]ServicePackage, len(servicePackages))
	for i := range servicePackages {
		pkgs[i] = PackageOf(servicePackages[i]...)
	}
	CreateContextualizedApplication(pkgs...).Join()
}

func CreateContextualizedApplication(servicePackages ...ServicePackage) Application {
	return startApplication(servicePackages)
}

// Deprecated: use CreateContextualizedApplication instead
func StartApplication(servicePackages ...[]any) Application {
	pkgs := make([]ServicePackage, len(servicePackages))
	for i := range servicePackages {
		pkgs[i] = PackageOf(servicePackages[i]...)
	}
	return CreateContextualizedApplication(pkgs...)
}

func startApplication(servicePackages []ServicePackage) Application {
	ctxInstance := func() *appContext {
		globalLock.Lock()
		defer globalLock.Unlock()

		ctxInstance := newApplicationContext()

		for _, pkg := range servicePackages {
			pkg.ForEach(func(service any, name string) {
				ctxInstance.register(service, name)
			})
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
		signal.Notify(osSignalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

		select {
		case sSig := <-osSignalCh:
			logger.Debug(ctxTag, "closing application by system signal:", sSig.String())
			break
		case <-app.stopCh:
			logger.Debug(ctxTag, "closing application by stop signal")
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

func GetTypedService[T any]() T {
	return GetService(u.GetInterfaceName[T]()).(T)
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
