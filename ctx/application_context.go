package ctx

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
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
	stopOrder []string

	stats  *appContextStats
	health *appContextHealth

	stopOnce sync.Once
}

func newApplicationContext() *appContext {
	root, cancel := context.WithCancel(context.Background())
	return &appContext{
		state:     stateNotInitialized,
		ctx:       root,
		ctxCancel: cancel,
		services:  make(map[string]*reflectiveServiceWrapper),
		states:    make(map[string]state),
		initOrder: make([]string, 0),
		stopOrder: make([]string, 0),
		stats:     createContextStats(),
		health:    createContextHealth(),
	}
}

func (ctx *appContext) register(serviceInstance any, name string) error {
	ctx.Lock()
	defer ctx.Unlock()

	if ctx.state != stateNotInitialized {
		return fmt.Errorf("wrong state: current (%s), expected (%s)", ctx.state.name, stateNotInitialized.name)
	}

	sInstance, err := newReflectiveServiceWrapper(ctx.ctx, serviceInstance, name)
	if err != nil {
		return err
	}

	serviceName := sInstance.Name()
	if _, found := ctx.services[serviceName]; found {
		return fmt.Errorf("service name duplication: [%s]", serviceName)
	}
	if serviceName == ctxTag {
		return fmt.Errorf("service can't have reserved name: [%s]", ctxTag)
	}
	ctx.services[serviceName] = sInstance
	ctx.states[serviceName] = stateNotInitialized
	logger.Debug(ctxTag, "registered service ["+serviceName+"] of", reflect.TypeOf(serviceInstance).String())
	return nil
}

// start initializes every service and publishes the initialized state. Lifecycle callbacks
// are deliberately invoked by afterStart after the global context has been published.
func (ctx *appContext) start() error {
	ctx.Lock()
	if ctx.state != stateNotInitialized {
		current := ctx.state
		ctx.Unlock()
		return fmt.Errorf("wrong state: current (%s), expected (%s)", current.name, stateNotInitialized.name)
	}
	ctx.state = stateInitialization
	names := make([]string, 0, len(ctx.services))
	for name := range ctx.services {
		names = append(names, name)
	}
	ctx.Unlock()
	sort.Strings(names)

	logger.Info(ctxTag, "=== starting... ===")
	for _, serviceName := range names {
		ctx.RLock()
		serviceState := ctx.states[serviceName]
		ctx.RUnlock()
		if serviceState != stateNotInitialized {
			continue
		}
		if err := ctx.initService(serviceName); err != nil {
			logger.Error(ctxTag, "on initialization ["+serviceName+"]:", err)
			ctx.cleanupFailedStart()
			return fmt.Errorf("can't initialize service [%s]: %w", serviceName, err)
		}
	}

	stopOrder, err := stableDependencyOrder(names, ctx.stats.dependencySnapshot())
	if err != nil {
		ctx.cleanupFailedStart()
		return fmt.Errorf("can't derive dependency-safe stop order: %w", err)
	}
	ctx.Lock()
	ctx.stopOrder = stopOrder
	ctx.state = stateInitialized
	ctx.Unlock()
	logger.Info(ctxTag, "=== all services have been initialized ===")
	return nil
}

func (ctx *appContext) afterStart() {
	services := ctx.serviceSnapshotByName()
	var wg sync.WaitGroup
	for _, named := range services {
		wg.Add(1)
		go func(serviceName string, serviceInstance *reflectiveServiceWrapper) {
			defer wg.Done()
			if err := serviceInstance.afterStart(); err != nil {
				logger.Error(ctxTag, "on service ["+serviceName+"] AfterStart():", err.Error())
			}
		}(named.name, named.service)
	}
	wg.Wait()
	logger.Info(ctxTag, "=== all lifecycle-aware services handled AfterStart event ===")
	logger.Info(ctxTag, "=== ...started ===")
}

func (ctx *appContext) stop() {
	ctx.stopOnce.Do(func() {
		ctx.RLock()
		if ctx.state != stateInitialized {
			ctx.RUnlock()
			return
		}
		services := make(map[string]*reflectiveServiceWrapper, len(ctx.services))
		for name, service := range ctx.services {
			services[name] = service
		}
		stopOrder := append([]string(nil), ctx.stopOrder...)
		ctx.RUnlock()

		logger.Info(ctxTag, "=== stopping... ===")
		for _, serviceName := range stopOrder {
			if err := services[serviceName].beforeStop(); err != nil {
				logger.Error(ctxTag, "on service ["+serviceName+"] BeforeStop():", err.Error())
			}
		}
		logger.Info(ctxTag, "=== all lifecycle-aware services handled BeforeStop event ===")

		ctx.Lock()
		ctx.state = stateUsed
		ctx.Unlock()
		logger.Debug(ctxTag, "canceling root context")
		ctx.ctxCancel()

		ctx.disposeServices(services)
		ctx.health.clear()

		ctx.Lock()
		ctx.services = nil
		ctx.states = nil
		ctx.Unlock()
		logger.Info(ctxTag, "=== ...stopped ===")
	})
}

func (ctx *appContext) GetService(serviceName string) (srv any, ok bool) {
	ctx.RLock()
	defer ctx.RUnlock()
	ctx.checkStateLocked(stateInitialized)

	service, found := ctx.services[serviceName]
	if !found {
		return nil, false
	}
	return service.unwrap(), true
}

func (ctx *appContext) Stats() AppContextStats {
	ctx.RLock()
	defer ctx.RUnlock()
	ctx.checkStateLocked(stateInitialized)
	return ctx.stats
}

func (ctx *appContext) Health() AppContextHealth {
	ctx.RLock()
	defer ctx.RUnlock()
	ctx.checkStateLocked(stateInitialized)
	return ctx.health
}

func (ctx *appContext) State() (int, string) {
	ctx.RLock()
	defer ctx.RUnlock()
	return ctx.state.code, ctx.state.name
}

func (ctx *appContext) initService(serviceName string) error {
	ctx.Lock()
	serviceState, found := ctx.states[serviceName]
	if !found {
		ctx.Unlock()
		return fmt.Errorf("service [%s] not found", serviceName)
	}
	switch serviceState {
	case stateInitialized:
		service := ctx.services[serviceName]
		ctx.Unlock()
		_ = service
		return nil
	case stateInitialization:
		ctx.Unlock()
		return fmt.Errorf("cyclic dependency involving [%s]", serviceName)
	case stateNotInitialized:
		ctx.states[serviceName] = stateInitialization
	default:
		ctx.Unlock()
		return fmt.Errorf("service [%s] is in unexpected state [%s]", serviceName, serviceState.name)
	}
	serviceInstance := ctx.services[serviceName]
	ctx.Unlock()

	logger.Debug(ctxTag, "service ["+serviceName+"] initialization started...")
	serviceDescriptor := createDescriptorFor(serviceInstance)
	provider := serviceProviderImpl(func(requestedServiceName string) any {
		logger.Debug(ctxTag, "["+serviceName+"] requested service ["+requestedServiceName+"]")
		if requestedServiceName == ctxTag {
			return ctx
		}

		serviceDescriptor.addDependency(requestedServiceName)
		ctx.RLock()
		requested, exists := ctx.services[requestedServiceName]
		ctx.RUnlock()
		if !exists {
			panic(fmt.Sprintf("service [%s] not found", requestedServiceName))
		}
		if err := ctx.initService(requestedServiceName); err != nil {
			panic(err)
		}
		return requested.unwrap()
	})

	var initErr error
	if panicErr := nopanic.Run(func() {
		initErr = serviceInstance.init(provider)
	}); panicErr != nil {
		initErr = panicErr.ToReasonError()
	}
	if initErr != nil {
		ctx.Lock()
		ctx.states[serviceName] = stateUsed
		ctx.Unlock()
		return initErr
	}

	ctx.Lock()
	ctx.states[serviceName] = stateInitialized
	ctx.initOrder = append(ctx.initOrder, serviceName)
	ctx.Unlock()
	ctx.stats.registerServiceDescriptor(serviceDescriptor)
	ctx.health.registerHealthReporter(serviceInstance)
	logger.Debug(ctxTag, "...service ["+serviceName+"] initialized")
	return nil
}

func (ctx *appContext) cleanupFailedStart() {
	ctx.Lock()
	ctx.state = stateUsed
	services := make(map[string]*reflectiveServiceWrapper, len(ctx.initOrder))
	for _, name := range ctx.initOrder {
		services[name] = ctx.services[name]
	}
	ctx.Unlock()
	ctx.ctxCancel()
	ctx.disposeServices(services)
	ctx.health.clear()
	ctx.Lock()
	ctx.services = nil
	ctx.states = nil
	ctx.Unlock()
}

func (ctx *appContext) disposeServices(services map[string]*reflectiveServiceWrapper) {
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)

	var wg sync.WaitGroup
	for _, serviceName := range names {
		serviceInstance := services[serviceName]
		wg.Add(1)
		go func(serviceName string, serviceInstance *reflectiveServiceWrapper) {
			defer wg.Done()
			logger.Debug(ctxTag, "dispose service ["+serviceName+"]")
			if err := serviceInstance.dispose(); err != nil {
				logger.Error(ctxTag, "on service ["+serviceName+"] disposing:", err.Error())
			}
			ctx.Lock()
			if ctx.states != nil {
				ctx.states[serviceName] = stateUsed
			}
			ctx.Unlock()
		}(serviceName, serviceInstance)
	}
	wg.Wait()
}

func (ctx *appContext) checkStateLocked(expectedState state) {
	if ctx.state != expectedState {
		logger.Fatal(ctxTag, "wrong state: current ("+ctx.state.name+"), expected ("+expectedState.name+")")
	}
}

type namedServiceSnapshot struct {
	name    string
	service *reflectiveServiceWrapper
}

func (ctx *appContext) serviceSnapshotByName() []namedServiceSnapshot {
	ctx.RLock()
	result := make([]namedServiceSnapshot, 0, len(ctx.services))
	for name, service := range ctx.services {
		result = append(result, namedServiceSnapshot{name: name, service: service})
	}
	ctx.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].name < result[j].name })
	return result
}
