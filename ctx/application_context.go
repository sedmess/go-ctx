package ctx

import (
	"log"
	"os"
	"os/signal"
	"strconv"
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

const contextDebugLog = "CTX_DEBUG"

const notInitialized = 0
const initialization = 1
const initialized = 2
const used = -1

type AppContext interface {
	Register(serviceInstance Service) AppContext
	Start()
	Stop()
	GetService(serviceName string) Service
}

type appContext struct {
	sync.RWMutex

	debug bool

	state int

	services map[string]Service
	states   map[string]int
}

var globalLock sync.Mutex
var applicationContextOnce sync.Once
var applicationContextInstance AppContext

func ApplicationContext() AppContext {
	applicationContextOnce.Do(func() {
		ctx := appContext{}
		ctx.debug = GetEnv(contextDebugLog).AsBoolDefault(false)
		ctx.state = notInitialized
		ctx.services = make(map[string]Service)
		ctx.states = make(map[string]int)
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

	ctx.checkState(notInitialized)

	serviceName := serviceInstance.Name()
	if _, found := ctx.services[serviceName]; found {
		log.Fatalln("CTX: ERR: service name duplication: [" + serviceName + "]")
	}
	ctx.services[serviceName] = serviceInstance
	ctx.states[serviceName] = notInitialized
	ctx.printLog("registered service [" + serviceName + "]")

	return ctx
}

func (ctx *appContext) Start() {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.checkState(notInitialized)
	ctx.state = initialization
	targetState := initialized

	for serviceName, serviceInstance := range ctx.services {
		if targetState == used {
			break
		}
		if ctx.states[serviceName] == notInitialized {
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println("CTX: ERR: on initialization ["+serviceName+"]:", err)
						ctx.disposeServices()
						targetState = used
					}
				}()
				ctx.initService(serviceInstance)
			}()
		}
	}

	if targetState == used {
		log.Fatalln("CTX: ERR: can't start context, see log above")
	}

	log.Println("CTX: started")

	for serviceName, serviceInstance := range ctx.services {
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			go func() {
				ctx.printLog("[" + serviceName + "] is livecycle awared, notify it for start event")
				lifecycleAwareInstance.AfterStart()
			}()
		}
	}

	ctx.state = targetState
}

func (ctx *appContext) Stop() {
	ctx.Lock()
	defer ctx.Unlock()

	if ctx.state != initialized {
		return
	}

	for serviceName, serviceInstance := range ctx.services {
		lifecycleAwareInstance, ok := serviceInstance.(LifecycleAware)
		if ok {
			go func() {
				ctx.printLog("[" + serviceName + "] is livecycle awared, notify it for stop event")
				lifecycleAwareInstance.BeforeStop()
			}()
		}
	}

	ctx.state = used

	ctx.disposeServices()

	ctx.services = nil
	ctx.states = nil

	log.Println("CTX: stopped")
}

func (ctx *appContext) GetService(serviceName string) Service {
	ctx.RLock()
	defer ctx.RUnlock()

	ctx.checkState(initialized)

	return ctx.services[serviceName]
}

func (ctx *appContext) initService(serviceInstance Service) {
	ctx.states[serviceInstance.Name()] = initialization
	ctx.printLog("service [" + serviceInstance.Name() + "] initialization started...")
	serviceInstance.Init(func(requestedServiceName string) Service {
		ctx.printLog("[" + serviceInstance.Name() + "] requested service [" + requestedServiceName + "]")
		if requestedServiceInstance, found := ctx.services[requestedServiceName]; found {
			serviceState := ctx.states[requestedServiceName]
			if serviceState == initialized {
				return requestedServiceInstance
			} else if serviceState == initialization {
				panic("CTX: ERR: cyclic dependency between [" + serviceInstance.Name() + "] and [" + requestedServiceName + "]")
				return nil
			} else if serviceState == notInitialized {
				ctx.initService(requestedServiceInstance)
				return requestedServiceInstance
			} else {
				panic("unexpected error")
			}
		} else {
			log.Fatalln("CTX: ERR: service [" + requestedServiceName + "] not found")
			return nil
		}
	})
	ctx.printLog("...service [" + serviceInstance.Name() + "] initialized")
	ctx.states[serviceInstance.Name()] = initialized
}

func (ctx *appContext) disposeServices() {
	for serviceName, serviceInstance := range ctx.services {
		if ctx.states[serviceName] == initialized {
			ctx.printLog("dispose service [" + serviceName + "]")
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println("CTX: ERR: on service ["+serviceName+"] disposing:", err)
					}
				}()
				serviceInstance.Dispose()
				ctx.states[serviceName] = used
			}()
		}
	}
}

func (ctx *appContext) checkState(expectedState int) {
	if ctx.state != expectedState {
		log.Panicln("CTX: ERR: wrong state: current (" + strconv.Itoa(ctx.state) + "), expected (" + strconv.Itoa(expectedState) + ")")
	}
}

func (ctx *appContext) printLog(msg string) {
	if ctx.debug {
		log.Println("CTX: " + msg)
	}
}
