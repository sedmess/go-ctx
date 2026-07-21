package ctx

import (
	"reflect"
	"sync"
	"testing"
)

func TestStatisticsServicesReturnsDeepSnapshot(t *testing.T) {
	stats := createContextStats()
	stats.registerServiceDescriptor(ServiceDescriptor{
		Name:         "consumer",
		Type:         reflect.TypeOf(&duplicateService{}),
		Dependencies: []string{"dependency"},
	})

	first := stats.Services()
	descriptor := first["consumer"]
	descriptor.Dependencies[0] = "mutated"
	first["consumer"] = descriptor
	delete(first, "consumer")

	second := stats.Services()
	if got := second["consumer"].Dependencies; !reflect.DeepEqual(got, []string{"dependency"}) {
		t.Fatalf("dependencies after caller mutation = %v", got)
	}
}

func TestStatisticsSnapshotsAreSafeForConcurrentConsumers(t *testing.T) {
	stats := createContextStats()
	stats.registerServiceDescriptor(ServiceDescriptor{Name: "service", Dependencies: []string{"one", "two"}})
	var wait sync.WaitGroup
	for i := 0; i < 64; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			services := stats.Services()
			descriptor := services["service"]
			descriptor.Dependencies[0] = "local"
		}()
	}
	wait.Wait()
	if got := stats.Services()["service"].Dependencies[0]; got != "one" {
		t.Fatalf("container dependency changed to %q", got)
	}
}
