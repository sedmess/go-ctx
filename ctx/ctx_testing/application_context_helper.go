package ctx_testing

import "github.com/sedmess/go-ctx/ctx"

type TestingApplication interface {
	WithTestingService()
	WithTestingServices()
	Start()
	Await()
}

type testingApplication struct {
	services        []any
	testingServices map[string]any
	app             ctx.Application
}

func CreateTestingApplication(servicePackages ...[]any) TestingApplication {
	services := make([]any, 0)
	for _, servicePackage := range servicePackages {
		services = append(services, servicePackage...)
	}
	return &testingApplication{services: services}
}

func (app *testingApplication) WithTestingService() {
	//TODO implement me
	panic("implement me")
}

func (app *testingApplication) WithTestingServices() {
	//TODO implement me
	panic("implement me")
}

func (app *testingApplication) Start() {
	//TODO implement me
	panic("implement me")
}

func (app *testingApplication) Await() {
	//TODO implement me
	panic("implement me")
}
