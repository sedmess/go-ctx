package ctx

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/sedmess/go-ctx/ctx/autoctx"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u"
)

var globalLock sync.RWMutex
var ctx *appContext
var startingContext *appContext
var notifySignals = signal.Notify
var stopSignalNotifications = signal.Stop

type Application interface {
	Stop() Application
	Join()
}

type application struct {
	stopOnce  sync.Once
	stopCh    chan struct{}
	startDone chan struct{}
	doneCh    chan struct{}
	signalCh  chan os.Signal
}

// Stop records a stop request. It is safe to call repeatedly or concurrently and does not
// wait for shutdown cleanup; call Join when completion is required.
func (a *application) Stop() Application {
	a.stopOnce.Do(func() { close(a.stopCh) })
	return a
}

// Join waits until lifecycle callbacks, cancellation, disposal, signal unregistration, and
// global-context cleanup have completed. Repeated and concurrent calls are safe.
func (a *application) Join() {
	<-a.doneCh
}

func CreateContextualizedApplication(servicePackages ...ServicePackage) Application {
	return startApplication(servicePackages)
}

func CreateAutoContextualizedApplication() Application {
	return startApplication([]ServicePackage{PackageOf(autoctx.RegisteredServices()...)})
}

func startApplication(servicePackages []ServicePackage) Application {
	InitSlog()
	ctxInstance := newApplicationContext()

	globalLock.Lock()
	if ctx != nil || startingContext != nil {
		globalLock.Unlock()
		panic("an application context is already active")
	}
	startingContext = ctxInstance
	globalLock.Unlock()

	clearGlobal := func() {
		globalLock.Lock()
		if ctx == ctxInstance {
			ctx = nil
		}
		if startingContext == ctxInstance {
			startingContext = nil
		}
		globalLock.Unlock()
	}

	var registrationErr error
	for _, pkg := range servicePackages {
		pkg.ForEach(func(service any, name string) {
			if registrationErr == nil {
				registrationErr = ctxInstance.register(service, name)
			}
		})
	}
	if registrationErr != nil {
		ctxInstance.cleanupFailedStart()
		clearGlobal()
		logger.Fatal(ctxTag, "can't register services:", registrationErr)
	}
	if err := ctxInstance.start(); err != nil {
		clearGlobal()
		logger.Fatal(ctxTag, "can't start context:", err)
	}
	globalLock.Lock()
	if startingContext != ctxInstance || ctx != nil {
		globalLock.Unlock()
		ctxInstance.cleanupFailedStart()
		panic("application context reservation was lost during startup")
	}
	ctx = ctxInstance
	startingContext = nil
	globalLock.Unlock()

	app := &application{
		stopCh:    make(chan struct{}),
		startDone: make(chan struct{}),
		doneCh:    make(chan struct{}),
		signalCh:  make(chan os.Signal, 1),
	}
	notifySignals(app.signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case received := <-app.signalCh:
			logger.Debug(ctxTag, "closing application by system signal:", received.String())
		case <-app.stopCh:
			logger.Debug(ctxTag, "closing application by stop signal")
		}

		stopSignalNotifications(app.signalCh)
		<-app.startDone
		ctxInstance.stop()
		clearGlobal()
		close(app.doneCh)
	}()

	ctxInstance.afterStart()
	close(app.startDone)
	return app
}

func GetService(serviceName string) (svc any, ok bool) {
	globalLock.RLock()
	current := ctx
	globalLock.RUnlock()
	if current == nil {
		panic("no active context")
	}
	return current.GetService(serviceName)
}

// GetTypedService returns the registered typed service. Ordinary absence returns the zero
// value and false; an incompatible value registered under the expected name is invalid
// wiring and causes a diagnostic panic.
func GetTypedService[T any]() (svc T, ok bool) {
	serviceName := u.GetInterfaceName[T]()
	srv, found := GetService(serviceName)
	if !found {
		var zero T
		return zero, false
	}
	typed, assignable := srv.(T)
	if !assignable {
		actualType := "<nil>"
		if actual := reflect.TypeOf(srv); actual != nil {
			actualType = actual.String()
		}
		panic(fmt.Sprintf("service [%s] has incompatible type: expected %s, actual %s", serviceName, reflect.TypeFor[T](), actualType))
	}
	return typed, true
}
