package ctx_testing

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/logger"
	"github.com/sedmess/go-ctx/u"
	"os"
	"reflect"
	"sync"
)

const ctxTag = "TESTING_CTX"

func init() {
	logger.Init(logger.DEBUG)
}

type TestingApplication interface {
	WithTestingService(service ctx.NamedService) TestingApplication
	WithTestingServices(pkg ctx.ServicePackage) TestingApplication
	WithParameter(key string, value string) TestingApplication
	Run(runFn func() int) int
}

type testingApplication struct {
	mu sync.Mutex

	params map[string]string

	basePackages    []ctx.ServicePackage
	testingPackages []ctx.ServicePackage
	app             ctx.Application
}

func CreateTestingApplication(servicePackages ...ctx.ServicePackage) TestingApplication {
	return &testingApplication{params: make(map[string]string), basePackages: servicePackages, testingPackages: make([]ctx.ServicePackage, 0)}
}

func Instead[T any](service T) ctx.NamedService {
	return ctx.WithName(u.GetInterfaceName[T](), service)
}

func (app *testingApplication) WithTestingService(service ctx.NamedService) TestingApplication {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.testingPackages = append(app.testingPackages, ctx.PackageOf(service))
	return app
}

func (app *testingApplication) WithTestingServices(pkg ctx.ServicePackage) TestingApplication {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.testingPackages = append(app.testingPackages, pkg)
	return app
}

func (app *testingApplication) WithParameter(key string, value string) TestingApplication {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.params[key] = value

	return app
}

func (app *testingApplication) Run(runFn func() int) int {
	app.mu.Lock()
	defer app.mu.Unlock()

	paramsBackup := make(map[string]string)
	for key, value := range app.params {
		currentValue := os.Getenv(key)
		if currentValue != "" {
			paramsBackup[key] = currentValue
			logger.Debug(ctxTag, "set env", key, ", old value stored")
		} else {
			logger.Debug(ctxTag, "set env", key)
		}
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range app.params {
			prevValue := paramsBackup[key]
			if prevValue != "" {
				logger.Debug(ctxTag, "restored env", key)
				_ = os.Setenv(key, prevValue)
			} else {
				_ = os.Setenv(key, "")
			}
		}
	}()

	serviceMap := make(map[string]any)

	for _, pkg := range app.basePackages {
		pkg.ForEach(func(service any, name string) {
			if name == "" {
				name = ctx.DefineServiceName(service)
			}
			serviceMap[name] = service
		})
	}
	for _, pkg := range app.testingPackages {
		pkg.ForEach(func(service any, name string) {
			if name == "" {
				name = ctx.DefineServiceName(service)
			}
			if bSvc, found := serviceMap[name]; found {
				logger.Debug(ctxTag, "substitute base service ["+name+"] of type", reflect.TypeOf(bSvc).String(), "with", reflect.TypeOf(service).String())
			}
			serviceMap[name] = service
		})
	}

	services := make([]any, 0)
	for name, service := range serviceMap {
		services = append(services, ctx.WithName(name, service))
	}

	app.app = ctx.CreateContextualizedApplication(ctx.PackageOf(services...))
	defer func() { app.app.Stop().Join() }()

	return runFn()
}
