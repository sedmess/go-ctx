package ctx

import (
	"math/rand"
	"testing"

	"github.com/sedmess/go-ctx/ctx/health"
)

type testHealthReporter struct {
	name   string
	status health.ServiceHealthStatus
}

func (reporter *testHealthReporter) Name() string { return reporter.name }
func (reporter *testHealthReporter) Health() health.ServiceHealth {
	return health.Status(reporter.status)
}

func aggregateStatusesForTest(t *testing.T, statuses []health.ServiceHealthStatus) health.ServiceHealth {
	t.Helper()
	aggregate := createContextHealth()
	for index, status := range statuses {
		reporter := &testHealthReporter{name: string(rune('a' + index)), status: status}
		wrapper, err := newReflectiveServiceWrapper(nil, reporter, "")
		if err != nil {
			t.Fatal(err)
		}
		aggregate.registerHealthReporter(wrapper)
	}
	return aggregate.Aggregate()
}

func TestHealthAggregationUsesWorstSeverityForEveryPermutation(t *testing.T) {
	base := []health.ServiceHealthStatus{health.Up, health.Partially, health.Down, health.DownCritical}
	random := rand.New(rand.NewSource(84))
	for i := 0; i < 1000; i++ {
		statuses := append([]health.ServiceHealthStatus(nil), base...)
		random.Shuffle(len(statuses), func(i, j int) { statuses[i], statuses[j] = statuses[j], statuses[i] })
		result := aggregateStatusesForTest(t, statuses)
		if result.Status != health.Down {
			t.Fatalf("iteration %d: aggregate = %s, want %s", i, result.Status, health.Down)
		}
	}
}

func TestHealthAggregationUnknownFailsSafeAndOwnsComponentMap(t *testing.T) {
	aggregate := createContextHealth()
	reporter := &testHealthReporter{name: "unknown", status: "UNKNOWN"}
	wrapper, err := newReflectiveServiceWrapper(nil, reporter, "")
	if err != nil {
		t.Fatal(err)
	}
	aggregate.registerHealthReporter(wrapper)
	result := aggregate.Aggregate()
	if result.Status != health.Down {
		t.Fatalf("unknown aggregate = %s, want %s", result.Status, health.Down)
	}
	delete(result.Components, "unknown")
	second := aggregate.Aggregate()
	if len(second.Components) != 1 {
		t.Fatal("caller mutation affected a later health result")
	}
}

func TestHealthAggregationEmptyAndNoncriticalDown(t *testing.T) {
	if got := createContextHealth().Aggregate().Status; got != health.Up {
		t.Fatalf("empty aggregate = %s, want %s", got, health.Up)
	}
	if got := aggregateStatusesForTest(t, []health.ServiceHealthStatus{health.Down}).Status; got != health.Partially {
		t.Fatalf("noncritical down aggregate = %s, want %s", got, health.Partially)
	}
}
