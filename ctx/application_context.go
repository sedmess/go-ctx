package ctx

import (
	"context"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
	"reflect"
	"sync"
)

type state struct {
	name string
	code int
}

var stateNotInitialized = state{code: 0, name: "not_initialized"}
var stateInitialization = state{code: 1, name: "initialization"}
var stateInitialized = state{code: 2, name: "initialized"}
var stateUsed = state{code: -1, name: "used"}

const ctxTag = "CTX"

type AppContext interface {
	GetService(serviceName string) (any, bool)
	Stats() AppContextStats
	Health() AppContextHealth
	State() (int, string)
}

type appContext struct {
	sync.RWMutex

	state state

	ctx       context.Context
	ctxCancel context.CancelFunc

	services  map[string]*reflectiveServiceWrapper
	states    map[string]state
	initOrder []string

	stats  *appContextStats
	health *appContextHealth

	eventBus chan event
}

func newApplicationContext() *appContext {
	ctx := appContext{}
	ctx.state = stateNotInitialized
	ctx.services = make(map[string]*reflectiveServiceWrapper)
	ctx.states = make(map[string]state)
	ctx.initOrder = make([]string, 0)
	ctx.eventBus = make(chan event)
	ctx.stats = createContextStats()
	ctx.health = createContextHealth()
	ctx.ctx, ctx.ctxCancel = context.WithCancel(context.Background())
	return &ctx
}

func (ctx *appContext) register(serviceInstance any, name string) {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(stateNotInitialized)

	sInstance := newReflectiveServiceWrapper(ctx.ctx, serviceInstance, name)

	serviceName := sInstance.Name()
	if _, found := ctx.services[serviceName]; found {
		logger.Fatal(ctxTag, "service name duplication: ["+serviceName+"]")
	}
	if serviceName == ctxTag {
		logger.Fatal(ctxTag, "service can't have reserved name: ["+ctxTag+"]")
	}
	ctx.services[serviceName] = sInstance
	ctx.states[serviceName] = stateNotInitialized
	logger.Debug(ctxTag, "registered service ["+serviceName+"] of", reflect.TypeOf(serviceInstance).String())
}

func (ctx *appContext) start() {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(stateNotInitialized)
	ctx.state = stateInitialization
	targetState := stateInitialized

	logger.Info(ctxTag, "=== starting... ===")

	for serviceName, serviceInstance := range ctx.services {
		if targetState == stateUsed {
			break
		}
		if ctx.states[serviceName] == stateNotInitialized {
			if err := nopanic.Run(func() {
				ctx.initService(serviceInstance)
			}); err != nil {
				logger.Error(ctxTag, "on initialization ["+serviceName+"]:", err.Error())

				ctx.disposeServices()
				targetState = stateUsed
			}
		}
	}

	if targetState == stateUsed {
		logger.Fatal(ctxTag, "can't start context, see log above")
	}

	logger.Info(ctxTag, "=== all services have been initialized ===")

	var wg sync.WaitGroup
	for serviceName, serviceInstance := range ctx.services {
		wg.Add(1)
		go func(serviceName string, serviceInstance *reflectiveServiceWrapper) {
			defer wg.Done()
			if err := serviceInstance.afterStart(); err != nil {
				logger.Error(ctxTag, "on service ["+serviceName+"] AfterStart():", err.Error())
			}
		}(serviceName, serviceInstance)
	}
	wg.Wait()

	logger.Info(ctxTag, "=== all lifecycle-aware services handled AfterStart event ===")

	logger.Info(ctxTag, "=== ...started ===")

	ctx.state = targetState
}

func (ctx *appContext) eventLoop(finishCh chan<- bool) {
	defer func() { finishCh <- true }()
	for {
		e := <-ctx.eventBus
		switch e.kind {
		case eUnhandledPanic:
			ctx.stop()
			panicPayload := e.payload.(panicPayload)
			logger.Fatal(ctxTag, "unhandled panic:", panicPayload.reason, "at\n", string(panicPayload.stack))
			return
		case eSuppressedPanic:
			panicPayload := e.payload.(panicPayload)
			logger.Error(ctxTag, "unhandled panic:", panicPayload.reason, "at\n", string(panicPayload.stack))
		case eStop:
			logger.Info(ctxTag, "stop application event received")
			ctx.stop()
			return
		}
	}
}

func (ctx *appContext) stop() {
	ctx.Lock()
	defer ctx.Unlock()

	if ctx.state != stateInitialized {
		return
	}

	logger.Info(ctxTag, "=== stopping... ===")

	for i := len(ctx.initOrder) - 1; i >= 0; i-- {
		serviceName := ctx.initOrder[i]

		serviceInstance := ctx.services[serviceName]

		if err := serviceInstance.beforeStop(); err != nil {
			logger.Error(ctxTag, "on service ["+serviceName+"] BeforeStop():", err.Error())
		}
	}

	logger.Info(ctxTag, "=== all lifecycle-aware services handled BeforeStop event ===")

	logger.Debug(ctxTag, "canceling root context")
	ctx.ctxCancel()

	ctx.state = stateUsed

	ctx.disposeServices()

	ctx.services = nil
	ctx.states = nil

	logger.Info(ctxTag, "=== ...stopped ===")
}

func (ctx *appContext) GetService(serviceName string) (srv any, ok bool) {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(stateInitialized)

	service, found := ctx.services[serviceName]
	if !found {
		return nil, false
	}
	return service.unwrap(), true
}

func (ctx *appContext) Stats() AppContextStats {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(stateInitialized)

	return ctx.stats
}

func (ctx *appContext) Health() AppContextHealth {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(stateInitialized)

	return ctx.health
}

func (ctx *appContext) State() (int, string) {
	state := ctx.state
	return state.code, state.name
}

func (ctx *appContext) initService(serviceInstance *reflectiveServiceWrapper) {
	ctx.states[serviceInstance.Name()] = stateInitialization
	logger.Debug(ctxTag, "service ["+serviceInstance.Name()+"] initialization started...")
	ctx.initOrder = append(ctx.initOrder, serviceInstance.Name())

	serviceDescriptor := createDescriptorFor(serviceInstance)
	ctx.health.registerHealthReporter(serviceInstance)

	if err := serviceInstance.init(serviceProviderImpl(func(requestedServiceName string) any {
		logger.Debug(ctxTag, "["+serviceInstance.Name()+"] requested service ["+requestedServiceName+"]")

		if requestedServiceName == ctxTag {
			return ctx
		}

		serviceDescriptor.addDependency(requestedServiceName)
		if requestedServiceInstance, found := ctx.services[requestedServiceName]; found {
			serviceState := ctx.states[requestedServiceName]
			if serviceState == stateInitialized {
				return requestedServiceInstance.unwrap()
			} else if serviceState == stateInitialization {
				logger.Fatal(ctxTag, "cyclic dependency between ["+serviceInstance.Name()+"] and ["+requestedServiceName+"]")
				return nil
			} else if serviceState == stateNotInitialized {
				ctx.initService(requestedServiceInstance)
				return requestedServiceInstance.unwrap()
			} else {
				panic("unexpected error")
			}
		} else {
			logger.Fatal(ctxTag, "service ["+requestedServiceName+"] not found")
			return nil
		}
	})); err != nil {

	}
	logger.Debug(ctxTag, "...service ["+serviceInstance.Name()+"] initialized")
	ctx.states[serviceInstance.Name()] = stateInitialized
	ctx.stats.registerServiceDescriptor(serviceDescriptor)
}

func (ctx *appContext) disposeServices() {
	var wg sync.WaitGroup
	var l sync.Mutex
	for serviceName, serviceInstance := range ctx.services {
		l.Lock()
		state := ctx.states[serviceName]
		l.Unlock()

		if state == stateInitialized {
			wg.Add(1)
			logger.Debug(ctxTag, "dispose service ["+serviceName+"]")
			go func(serviceName string, serviceInstance *reflectiveServiceWrapper) {
				defer wg.Done()

				if err := serviceInstance.dispose(); err != nil {
					logger.Error(ctxTag, "on service ["+serviceName+"] disposing:", err.Error())
				}
				l.Lock()
				ctx.states[serviceName] = stateUsed
				l.Unlock()
			}(serviceName, serviceInstance)
		}
	}
	wg.Wait()
}

func (ctx *appContext) checkState(expectedState state) {
	if ctx.state != expectedState {
		logger.Fatal(ctxTag, "wrong state: current ("+ctx.state.name+"), expected ("+expectedState.name+")")
	}
}

func (ctx *appContext) sendEvent(e event) {
	ctx.eventBus <- e
}
