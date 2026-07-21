package ctx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type lifecycleAccessProbe struct {
	app  AppContext `ctx:"CTX"`
	err  chan error
	seen atomic.Int32
}

func (probe *lifecycleAccessProbe) check(phase string) {
	code, name := probe.app.State()
	if code != stateInitialized.code || name != stateInitialized.name {
		probe.err <- fmt.Errorf("%s state = (%d, %s)", phase, code, name)
		return
	}
	if _, found := GetService("probe"); !found {
		probe.err <- fmt.Errorf("%s could not look up service", phase)
		return
	}
	if _, found := probe.app.Stats().Services()["probe"]; !found {
		probe.err <- fmt.Errorf("%s could not read statistics", phase)
		return
	}
	_ = probe.app.Health().Aggregate()
	probe.seen.Add(1)
}

func (probe *lifecycleAccessProbe) AfterStart() { probe.check("AfterStart") }
func (probe *lifecycleAccessProbe) BeforeStop() { probe.check("BeforeStop") }

func TestLifecycleCallbacksCanReenterContext(t *testing.T) {
	probe := &lifecycleAccessProbe{err: make(chan error, 2)}
	app := CreateContextualizedApplication(PackageOf(WithName("probe", probe)))
	app.Stop().Join()

	close(probe.err)
	for err := range probe.err {
		t.Error(err)
	}
	if got := probe.seen.Load(); got != 2 {
		t.Fatalf("callback checks = %d, want 2", got)
	}
}

func TestApplicationConcurrentStopJoinAndStateReads(t *testing.T) {
	probe := &lifecycleAccessProbe{err: make(chan error, 2)}
	app := CreateContextualizedApplication(PackageOf(WithName("probe", probe)))

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for i := 0; i < 1000; i++ {
			probe.app.State()
		}
	}()

	var callers sync.WaitGroup
	for i := 0; i < 64; i++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			app.Stop()
			app.Join()
		}()
	}
	callers.Wait()
	<-readerDone
	app.Stop().Join()
	if code, name := probe.app.State(); code != stateUsed.code || name != stateUsed.name {
		t.Fatalf("terminal state = (%d, %s)", code, name)
	}
}

type restartProbe struct{ stopped atomic.Int32 }

func (probe *restartProbe) BeforeStop() { probe.stopped.Add(1) }

func TestApplicationRestartCycles(t *testing.T) {
	probe := &restartProbe{}
	for i := 0; i < 100; i++ {
		app := CreateContextualizedApplication(PackageOf(probe))
		app.Stop().Stop().Join()
	}
	if got := probe.stopped.Load(); got != 100 {
		t.Fatalf("stop callbacks = %d, want 100", got)
	}
}

type lifecycleFailureProbe struct {
	disposed atomic.Bool
}

func (*lifecycleFailureProbe) AfterStart()    { panic("start callback") }
func (*lifecycleFailureProbe) BeforeStop()    { panic("stop callback") }
func (probe *lifecycleFailureProbe) Dispose() { probe.disposed.Store(true) }

func TestLifecycleCallbackFailureDoesNotSkipCleanup(t *testing.T) {
	probe := &lifecycleFailureProbe{}
	app := CreateContextualizedApplication(PackageOf(probe))
	app.Stop().Join()
	if !probe.disposed.Load() {
		t.Fatal("service was not disposed after callback failures")
	}
}

type startupDependency struct {
	root     context.Context `ctx:""`
	disposed atomic.Bool
	canceled atomic.Bool
}

func (dependency *startupDependency) Dispose() {
	dependency.disposed.Store(true)
	dependency.canceled.Store(dependency.root.Err() != nil)
}

type failingStartupService struct {
	dependency *startupDependency `ctx:""`
}

func (*failingStartupService) Init() error { return errors.New("expected startup failure") }

func TestStartupFailureDisposesInitializedDependencies(t *testing.T) {
	dependency := &startupDependency{}
	context := newApplicationContext()
	if err := context.register(dependency, ""); err != nil {
		t.Fatal(err)
	}
	if err := context.register(&failingStartupService{}, ""); err != nil {
		t.Fatal(err)
	}
	if err := context.start(); err == nil {
		t.Fatal("start succeeded; want failure")
	}
	if !dependency.disposed.Load() {
		t.Fatal("initialized dependency was not disposed")
	}
	if !dependency.canceled.Load() {
		t.Fatal("root context was not canceled before dependency disposal")
	}
	if code, _ := context.State(); code != stateUsed.code {
		t.Fatalf("state code = %d, want %d", code, stateUsed.code)
	}
}

type blockingStopService struct {
	app     Application
	entered chan struct{}
	release chan struct{}
	count   atomic.Int32
}

func (service *blockingStopService) BeforeStop() {
	service.count.Add(1)
	service.app.Stop()
	close(service.entered)
	<-service.release
}

func TestApplicationStopReturnsBeforeCleanupAndIsReentrant(t *testing.T) {
	service := &blockingStopService{entered: make(chan struct{}), release: make(chan struct{})}
	app := CreateContextualizedApplication(PackageOf(service))
	service.app = app

	stopReturned := make(chan struct{})
	go func() {
		app.Stop()
		close(stopReturned)
	}()
	waitForTest(t, stopReturned, "Stop blocked waiting for the shutdown coordinator")
	waitForTest(t, service.entered, "BeforeStop did not begin")
	close(service.release)
	app.Join()
	app.Stop().Join()
	if got := service.count.Load(); got != 1 {
		t.Fatalf("BeforeStop calls = %d, want 1", got)
	}
}

type unsupportedConfigurationService struct {
	Value complex64 `env:"GO_CTX_UNSUPPORTED"`
}

func TestStartupConfigurationFailurePropagatesInternally(t *testing.T) {
	context := newApplicationContext()
	if err := context.register(&unsupportedConfigurationService{}, ""); err != nil {
		t.Fatal(err)
	}
	if err := context.start(); err == nil || !strings.Contains(err.Error(), "type not supported") {
		t.Fatalf("start error = %v", err)
	}
	if code, _ := context.State(); code != stateUsed.code {
		t.Fatalf("state code = %d, want %d", code, stateUsed.code)
	}
}

func waitForTest(t *testing.T, ch <-chan struct{}, message string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal(message)
	}
}
