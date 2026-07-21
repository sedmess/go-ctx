package ctx

import (
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestStableDependencyOrderAcrossPermutations(t *testing.T) {
	dependencies := map[string][]string{
		"api":    {"cache", "store"},
		"cache":  {"store"},
		"worker": {"store"},
	}
	want := []string{"api", "cache", "worker", "store"}
	random := rand.New(rand.NewSource(42))
	base := []string{"store", "worker", "api", "cache"}
	for i := 0; i < 1000; i++ {
		names := append([]string(nil), base...)
		random.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })
		got, err := stableDependencyOrder(names, dependencies)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("iteration %d: order = %v, want %v", i, got, want)
		}
	}
}

type stopOrderRecorder struct {
	mu    sync.Mutex
	names []string
}

func (recorder *stopOrderRecorder) add(name string) {
	recorder.mu.Lock()
	recorder.names = append(recorder.names, name)
	recorder.mu.Unlock()
}

type orderedDependency struct{ recorder *stopOrderRecorder }

func (service *orderedDependency) BeforeStop() { service.recorder.add("dependency") }

type orderedConsumer struct {
	dependency *orderedDependency `ctx:"dependency"`
	recorder   *stopOrderRecorder
}

func (service *orderedConsumer) BeforeStop() { service.recorder.add("consumer") }

func TestApplicationStopsConsumerBeforeDependency(t *testing.T) {
	recorder := &stopOrderRecorder{}
	dependency := &orderedDependency{recorder: recorder}
	consumer := &orderedConsumer{recorder: recorder}
	app := CreateContextualizedApplication(PackageOf(
		WithName("dependency", dependency),
		WithName("consumer", consumer),
	))
	app.Stop().Join()

	want := []string{"consumer", "dependency"}
	if !reflect.DeepEqual(recorder.names, want) {
		t.Fatalf("stop order = %v, want %v", recorder.names, want)
	}
}
