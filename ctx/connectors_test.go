package ctx

import (
	"sync"
	"testing"
	"time"
)

type connectorA struct {
	ServiceConnector[string, int]
	received chan string
}

func newConnectorA() *connectorA {
	service := &connectorA{received: make(chan string, 2)}
	service.ServiceConnector = NewServiceConnector[string, int](service)
	return service
}

func (*connectorA) Name() string                   { return "a" }
func (*connectorA) Init(ServiceProvider)           {}
func (*connectorA) Dispose()                       {}
func (service *connectorA) OnMessage(value string) { service.received <- value }

type connectorB struct {
	ServiceConnector[int, string]
	received chan int
	entered  chan struct{}
	release  <-chan struct{}
	stopping chan struct{}
	once     sync.Once
}

func newConnectorB() *connectorB {
	service := &connectorB{received: make(chan int, 2)}
	service.ServiceConnector = NewServiceConnector[int, string](service)
	return service
}

func (*connectorB) Name() string         { return "b" }
func (*connectorB) Init(ServiceProvider) {}
func (*connectorB) Dispose()             {}
func (service *connectorB) BeforeStop() {
	if service.stopping != nil {
		close(service.stopping)
	}
	service.ServiceConnector.BeforeStop()
}
func (service *connectorB) OnMessage(value int) {
	service.received <- value
	if service.entered != nil {
		service.once.Do(func() { close(service.entered) })
		<-service.release
	}
}

func TestConnectorDeliveryStopAndRestart(t *testing.T) {
	a := newConnectorA()
	b := newConnectorB()
	start := func() Application {
		return CreateContextualizedApplication(PackageOf(a, b, ConnectServices("a", "b")))
	}

	app := start()
	a.Send(7)
	select {
	case got := <-b.received:
		if got != 7 {
			t.Fatalf("received %d, want 7", got)
		}
	case <-time.After(time.Second):
		t.Fatal("connector did not deliver")
	}
	app.Stop().Join()

	blockedSend := make(chan struct{})
	go func() {
		a.Send(8)
		close(blockedSend)
	}()
	waitForTest(t, blockedSend, "send remained blocked after shutdown")

	app = start()
	a.Send(9)
	select {
	case got := <-b.received:
		if got != 9 {
			t.Fatalf("received %d after restart, want 9", got)
		}
	case <-time.After(time.Second):
		t.Fatal("connector did not deliver after restart")
	}
	app.Stop().Join()
}

func TestConnectorStopWaitsForInflightHandler(t *testing.T) {
	a := newConnectorA()
	b := newConnectorB()
	b.entered = make(chan struct{})
	release := make(chan struct{})
	b.release = release
	b.stopping = make(chan struct{})
	app := CreateContextualizedApplication(PackageOf(a, b, ConnectServices("a", "b")))

	go a.Send(1)
	<-b.entered
	joined := make(chan struct{})
	go func() {
		app.Stop().Join()
		close(joined)
	}()
	<-b.stopping
	select {
	case <-joined:
		t.Fatal("application joined before the in-flight handler returned")
	default:
	}
	close(release)
	waitForTest(t, joined, "connector listener did not acknowledge exit")
}
