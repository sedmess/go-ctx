package ctx_testing

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/autoctx"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u"
	"os"
	"reflect"
	"sort"
	"sync"
)

const ctxTag = "TESTING_CTX"

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

func CreateAutoTestingApplication() TestingApplication {
	return &testingApplication{params: make(map[string]string), basePackages: []ctx.ServicePackage{ctx.PackageOf(autoctx.RegisteredServices()...)}, testingPackages: make([]ctx.ServicePackage, 0)}
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
	return withParameters(app.params, func() int {
		services := mergeServicePackages(app.basePackages, app.testingPackages)
		app.app = ctx.CreateContextualizedApplication(ctx.PackageOf(services...))
		defer func() { app.app.Stop().Join() }()

		return runFn()
	})
}

func withParameters(params map[string]string, runFn func() int) int {
	type envBackup struct {
		value   string
		present bool
	}
	paramsBackup := make(map[string]envBackup)
	for key := range params {
		currentValue, present := os.LookupEnv(key)
		paramsBackup[key] = envBackup{value: currentValue, present: present}
	}
	defer func() {
		for key := range params {
			previous := paramsBackup[key]
			if previous.present {
				logger.Debug(ctxTag, "restored env", key)
				_ = os.Setenv(key, previous.value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}()
	for key, value := range params {
		if paramsBackup[key].present {
			logger.Debug(ctxTag, "set env", key, ", old value stored")
		} else {
			logger.Debug(ctxTag, "set env", key)
		}
		_ = os.Setenv(key, value)
	}
	return runFn()
}

func mergeServicePackages(basePackages, testingPackages []ctx.ServicePackage) []any {
	serviceMap := make(map[string]any)
	testingNames := make(map[string]struct{})

	for _, pkg := range basePackages {
		pkg.ForEach(func(service any, name string) {
			if name == "" {
				name = ctx.DefineServiceName(service)
			}
			if _, duplicate := serviceMap[name]; duplicate {
				panic("duplicate base service [" + name + "]")
			}
			serviceMap[name] = service
		})
	}
	for _, pkg := range testingPackages {
		pkg.ForEach(func(service any, name string) {
			explicitName := name != ""
			if name == "" {
				name = ctx.DefineServiceName(service)
			}
			if _, duplicate := testingNames[name]; duplicate {
				panic("duplicate testing service [" + name + "]")
			}
			testingNames[name] = struct{}{}
			if bSvc, found := serviceMap[name]; found {
				if !explicitName {
					panic("testing service [" + name + "] must use an explicit substitution name")
				}
				logger.Debug(ctxTag, "substitute base service ["+name+"] of type", reflect.TypeOf(bSvc).String(), "with", reflect.TypeOf(service).String())
			}
			serviceMap[name] = service
		})
	}

	names := make([]string, 0, len(serviceMap))
	for name := range serviceMap {
		names = append(names, name)
	}
	sort.Strings(names)
	services := make([]any, 0, len(names))
	for _, name := range names {
		service := serviceMap[name]
		services = append(services, ctx.WithName(name, service))
	}

	return services
}
