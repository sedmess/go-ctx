package ctx

import (
	"os"
	"testing"
)

type signalTestService struct{}

func TestApplicationSignalAndExplicitStopShareCoordinator(t *testing.T) {
	originalNotify := notifySignals
	originalStop := stopSignalNotifications
	registered := make(chan chan<- os.Signal, 2)
	unregistered := make(chan chan<- os.Signal, 2)
	notifySignals = func(channel chan<- os.Signal, _ ...os.Signal) { registered <- channel }
	stopSignalNotifications = func(channel chan<- os.Signal) { unregistered <- channel }
	defer func() {
		notifySignals = originalNotify
		stopSignalNotifications = originalStop
	}()

	app := CreateContextualizedApplication(PackageOf(&signalTestService{})).(*application)
	firstSignalChannel := <-registered
	if capacity := cap(firstSignalChannel); capacity < 1 {
		t.Fatalf("signal channel capacity = %d, want at least 1", capacity)
	}
	firstSignalChannel <- os.Interrupt
	app.Stop()
	app.Join()
	app.Join()
	if stopped := <-unregistered; stopped != firstSignalChannel {
		t.Fatal("coordinator unregistered a different signal channel")
	}

	restarted := CreateContextualizedApplication(PackageOf(&signalTestService{}))
	secondSignalChannel := <-registered
	if secondSignalChannel == firstSignalChannel {
		t.Fatal("restart reused the prior signal channel")
	}
	restarted.Stop().Join()
	if stopped := <-unregistered; stopped != secondSignalChannel {
		t.Fatal("restart did not unregister its signal channel")
	}
	select {
	case extra := <-unregistered:
		t.Fatalf("signal channel unregistered more than once: %v", extra)
	default:
	}
}
