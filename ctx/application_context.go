package ctx

import (
	"github.com/sedmess/go-ctx/logger"
	"runtime/debug"
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
	GetService(serviceName string) any
	Stats() AppContextStats
	Health() AppContextHealth
	State() (int, string)
}

type appContext struct {
	sync.RWMutex

	state state

	services  map[string]Service
	states    map[string]state
	initOrder []string

	stats  *appContextStats
	health *appContextHealth

	eventBus chan event
}

func newApplicationContext() *appContext {
	ctx := appContext{}
	ctx.state = stateNotInitialized
	ctx.services = make(map[string]Service)
	ctx.states = make(map[string]state)
	ctx.initOrder = make([]string, 0)
	ctx.eventBus = make(chan event)
	ctx.stats = createContextStats()
	ctx.health = createContextHealth()
	return &ctx
}

func (ctx *appContext) register(serviceInstance any) {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(stateNotInitialized)

	sInstance := newReflectiveServiceWrapper(serviceInstance)

	serviceName := sInstance.Name()
	if _, found := ctx.services[serviceName]; found {
		logger.Fatal(ctxTag, "service name duplication: ["+serviceName+"]")
	}
	if serviceName == ctxTag {
		logger.Fatal(ctxTag, "service can't have reserved name: ["+ctxTag+"]")
	}
	ctx.services[serviceName] = sInstance
	ctx.states[serviceName] = stateNotInitialized
	logger.Debug(ctxTag, "registered service ["+serviceName+"]")
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
			runWithRecover(
				func() {
					ctx.initService(serviceInstance)
				},
				func(reason any) {
					logger.Error(ctxTag, "on initialization ["+serviceName+"]:", reason, "stacktrace:", string(debug.Stack()))

					ctx.disposeServices()
					targetState = stateUsed
				},
			)
		}
	}

	if targetState == stateUsed {
		logger.Fatal(ctxTag, "can't start context, see log above")
	}

	logger.Info(ctxTag, "=== all services have been initialized ===")

	var wg sync.WaitGroup
	hasLCServices := false
	for serviceName, serviceInstance := range ctx.services {
		serviceInstance := unwrap(serviceInstance)
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			hasLCServices = true
			wg.Add(1)
			go func(serviceName string) {
				defer wg.Done()
				logger.Debug(ctxTag, "["+serviceName+"] is livecycle-aware, notify it for start event")
				runWithRecover(
					lifecycleAwareInstance.AfterStart,
					func(reason any) {
						logger.Error(ctxTag, "on service ["+serviceName+"] AfterStart():", reason, "stacktrace:", string(debug.Stack()))
					},
				)
			}(serviceName)
		}
	}
	wg.Wait()

	if hasLCServices {
		logger.Info(ctxTag, "=== all lifecycle-aware services handled AfterStart event ===")
	}

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

	hasLCServices := false
	for i := len(ctx.initOrder) - 1; i >= 0; i-- {
		serviceName := ctx.initOrder[i]
		serviceInstance := unwrap(ctx.services[serviceName])
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			hasLCServices = true
			logger.Debug(ctxTag, "["+serviceName+"] is livecycle-aware, notify it for stop event")
			runWithRecover(
				func() {
					lifecycleAwareInstance.BeforeStop()
				},
				func(reason any) {
					logger.Error(ctxTag, "on service ["+serviceName+"] BeforeStop()", reason, "stacktrace:", string(debug.Stack()))
				},
			)
		}
	}

	if hasLCServices {
		logger.Info(ctxTag, "=== all lifecycle-aware services handled BeforeStop event ===")
	}

	ctx.state = stateUsed

	ctx.disposeServices()

	ctx.services = nil
	ctx.states = nil

	logger.Info(ctxTag, "=== ...stopped ===")
}

func (ctx *appContext) GetService(serviceName string) any {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(stateInitialized)

	return unwrap(ctx.services[serviceName])
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

func (ctx *appContext) initService(serviceInstance Service) {
	ctx.states[serviceInstance.Name()] = stateInitialization
	logger.Debug(ctxTag, "service ["+serviceInstance.Name()+"] initialization started...")
	ctx.initOrder = append(ctx.initOrder, serviceInstance.Name())

	serviceDescriptor := createDescriptorFor(serviceInstance)
	ctx.health.registerHealthReporter(serviceInstance)

	serviceInstance.Init(serviceProviderImpl(func(requestedServiceName string) any {
		logger.Debug(ctxTag, "["+serviceInstance.Name()+"] requested service ["+requestedServiceName+"]")

		if requestedServiceName == ctxTag {
			return ctx
		}

		serviceDescriptor.addDependency(requestedServiceName)
		if requestedServiceInstance, found := ctx.services[requestedServiceName]; found {
			serviceState := ctx.states[requestedServiceName]
			if serviceState == stateInitialized {
				return unwrap(requestedServiceInstance)
			} else if serviceState == stateInitialization {
				panic("CTX: ERR: cyclic dependency between [" + serviceInstance.Name() + "] and [" + requestedServiceName + "]")
				return nil
			} else if serviceState == stateNotInitialized {
				ctx.initService(requestedServiceInstance)
				return unwrap(requestedServiceInstance)
			} else {
				panic("unexpected error")
			}
		} else {
			logger.Fatal(ctxTag, "service ["+requestedServiceName+"] not found")
			return nil
		}
	}))
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
			go func(serviceName string) {
				defer wg.Done()
				runWithRecover(
					func() {
						serviceInstance.Dispose()

						l.Lock()
						ctx.states[serviceName] = stateUsed
						l.Unlock()
					},
					func(reason any) {
						logger.Error(ctxTag, "on service ["+serviceName+"] disposing:", reason, "stacktrace:", string(debug.Stack()))
					},
				)
			}(serviceName)
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
