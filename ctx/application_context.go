package ctx

import (
	"os"
	"os/signal"
	"sync"
)

type Service interface {
	Init(serviceProvider func(serviceName string) Service)
	Name() string
	Dispose()
}

type LifecycleAware interface {
	AfterStart()
	BeforeStop()
}

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
	Register(serviceInstance Service) AppContext
	Start()
	Stop()
	GetService(serviceName string) Service
}

type appContext struct {
	sync.RWMutex

	state state

	services map[string]Service
	states   map[string]state
}

var globalLock sync.Mutex
var applicationContextOnce sync.Once
var applicationContextInstance AppContext

func ApplicationContext() AppContext {
	applicationContextOnce.Do(func() {
		ctx := appContext{}
		ctx.state = stateNotInitialized
		ctx.services = make(map[string]Service)
		ctx.states = make(map[string]state)
		applicationContextInstance = &ctx
	})
	return applicationContextInstance
}

func StartContextualizedApplication(packageServices ...[]Service) {
	globalLock.Lock()
	defer globalLock.Unlock()

	ctxInstance := ApplicationContext()
	for _, services := range packageServices {
		for _, service := range services {
			ctxInstance.Register(service)
		}
	}

	defer ctxInstance.Stop()
	ctxInstance.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		ctxInstance.Stop()
		os.Exit(0)
	}()

	<-make(chan bool)
}

func (ctx *appContext) Register(serviceInstance Service) AppContext {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(stateNotInitialized)

	serviceName := serviceInstance.Name()
	if _, found := ctx.services[serviceName]; found {
		LogFatal(ctxTag, "service name duplication: ["+serviceName+"]")
	}
	ctx.services[serviceName] = serviceInstance
	ctx.states[serviceName] = stateNotInitialized
	LogDebug(ctxTag, "registered service ["+serviceName+"]")

	return ctx
}

func (ctx *appContext) Start() {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(stateNotInitialized)
	ctx.state = stateInitialization
	targetState := stateInitialized

	for serviceName, serviceInstance := range ctx.services {
		if targetState == stateUsed {
			break
		}
		if ctx.states[serviceName] == stateNotInitialized {
			func() {
				defer func() {
					if err := recover(); err != nil {
						LogError(ctxTag, "on initialization ["+serviceName+"]:", err)
						ctx.disposeServices()
						targetState = stateUsed
					}
				}()
				ctx.initService(serviceInstance)
			}()
		}
	}

	if targetState == stateUsed {
		LogFatal("can't start context, see log above")
	}

	LogInfo("started")

	for serviceName, serviceInstance := range ctx.services {
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			go func() {
				LogDebug(ctxTag, "["+serviceName+"] is livecycle-aware, notify it for start event")
				lifecycleAwareInstance.AfterStart()
			}()
		}
	}

	ctx.state = targetState
}

func (ctx *appContext) Stop() {
	ctx.Lock()
	defer ctx.Unlock()

	if ctx.state != stateInitialized {
		return
	}

	for serviceName, serviceInstance := range ctx.services {
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			go func() {
				LogDebug(ctxTag, "["+serviceName+"] is livecycle-aware, notify it for stop event")
				lifecycleAwareInstance.BeforeStop()
			}()
		}
	}

	ctx.state = stateUsed

	ctx.disposeServices()

	ctx.services = nil
	ctx.states = nil

	LogInfo(ctxTag, "stopped")
}

func (ctx *appContext) GetService(serviceName string) Service {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(stateInitialized)

	return ctx.services[serviceName]
}

func (ctx *appContext) initService(serviceInstance Service) {
	ctx.states[serviceInstance.Name()] = stateInitialization
	LogDebug(ctxTag, "service ["+serviceInstance.Name()+"] initialization started...")
	serviceInstance.Init(func(requestedServiceName string) Service {
		LogDebug(ctxTag, "["+serviceInstance.Name()+"] requested service ["+requestedServiceName+"]")
		if requestedServiceInstance, found := ctx.services[requestedServiceName]; found {
			serviceState := ctx.states[requestedServiceName]
			if serviceState == stateInitialized {
				return requestedServiceInstance
			} else if serviceState == stateInitialization {
				panic("CTX: ERR: cyclic dependency between [" + serviceInstance.Name() + "] and [" + requestedServiceName + "]")
				return nil
			} else if serviceState == stateNotInitialized {
				ctx.initService(requestedServiceInstance)
				return requestedServiceInstance
			} else {
				panic("unexpected error")
			}
		} else {
			LogFatal(ctxTag, "service ["+requestedServiceName+"] not found")
			return nil
		}
	})
	LogDebug(ctxTag, "...service ["+serviceInstance.Name()+"] initialized")
	ctx.states[serviceInstance.Name()] = stateInitialized
}

func (ctx *appContext) disposeServices() {
	for serviceName, serviceInstance := range ctx.services {
		if ctx.states[serviceName] == stateInitialized {
			LogDebug(ctxTag, "dispose service ["+serviceName+"]")
			func() {
				defer func() {
					if err := recover(); err != nil {
						LogError(ctxTag, "on service ["+serviceName+"] disposing:", err)
					}
				}()
				serviceInstance.Dispose()
				ctx.states[serviceName] = stateUsed
			}()
		}
	}
}

func (ctx *appContext) checkState(expectedState state) {
	if ctx.state != expectedState {
		LogFatal(ctxTag, "wrong state: current ("+ctx.state.name+"), expected ("+expectedState.name+")")
	}
}
