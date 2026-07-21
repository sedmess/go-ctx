package ctx

import (
	"fmt"
	"sort"
)

// stableDependencyOrder returns consumers before the dependencies they requested. When
// several services are ready at once, names provide a deterministic tie-breaker.
func stableDependencyOrder(names []string, dependencies map[string][]string) ([]string, error) {
	nodes := make(map[string]struct{}, len(names))
	for _, name := range names {
		nodes[name] = struct{}{}
	}

	indegree := make(map[string]int, len(nodes))
	edges := make(map[string][]string, len(nodes))
	for name := range nodes {
		indegree[name] = 0
	}
	for consumer, requested := range dependencies {
		if _, exists := nodes[consumer]; !exists {
			continue
		}
		seen := make(map[string]struct{}, len(requested))
		for _, dependency := range requested {
			if dependency == ctxTag {
				continue
			}
			if _, exists := nodes[dependency]; !exists {
				continue
			}
			if _, duplicate := seen[dependency]; duplicate {
				continue
			}
			seen[dependency] = struct{}{}
			edges[consumer] = append(edges[consumer], dependency)
			indegree[dependency]++
		}
		sort.Strings(edges[consumer])
	}

	ready := make([]string, 0, len(nodes))
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)

	result := make([]string, 0, len(nodes))
	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]
		result = append(result, name)
		for _, dependency := range edges[name] {
			indegree[dependency]--
			if indegree[dependency] == 0 {
				ready = append(ready, dependency)
				sort.Strings(ready)
			}
		}
	}
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("dependency graph contains a cycle")
	}
	return result, nil
}
