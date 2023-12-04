package ctxtesting

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/logger"
)

type contextAware struct {
	ctx   ctx.AppContext `inject:"CTX"`
	state chan bool
}

func (c *contextAware) AfterStart() {
	go func() {
		c.state <- true
	}()
}

func (c *contextAware) BeforeStop() {
}

type TestingApplication interface {
	WithServices(services ...any) TestingApplication
	WithServicePackages(servicePackages ...[]any) TestingApplication
	Start()
	Stop()
}

type testingApplication struct {
	contextAware *contextAware
	services     []any
	stopCh       chan bool
}

func TestingContextualizedApplication() TestingApplication {
	return &testingApplication{contextAware: &contextAware{state: make(chan bool)}, services: make([]any, 0), stopCh: make(chan bool)}
}

func (t *testingApplication) WithServices(services ...any) TestingApplication {
	for _, service := range services {
		t.services = append(t.services, service)
	}
	return t
}

func (t *testingApplication) WithServicePackages(servicePackages ...[]any) TestingApplication {
	for _, services := range servicePackages {
		for _, service := range services {
			t.services = append(t.services, service)
		}
	}
	return t
}

func (t *testingApplication) Start() {
	logger.Init(logger.DEBUG)
	go func() {
		ctx.StartContextualizedApplication(ctx.ServiceArray(t.contextAware), ctx.ServiceArray(t.services...))
		t.stopCh <- true
	}()
	<-t.contextAware.state
}

func (t *testingApplication) Stop() {
	ctx.StopContext()
	<-t.stopCh
}
