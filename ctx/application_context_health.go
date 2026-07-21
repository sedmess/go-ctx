package ctx

import (
	"sync"

	"github.com/sedmess/go-ctx/ctx/health"
)

type AppContextHealth interface {
	Aggregate() health.ServiceHealth
}

type appContextHealth struct {
	mu       sync.RWMutex
	services map[string]HealthReporter
}

func createContextHealth() *appContextHealth {
	return &appContextHealth{services: make(map[string]HealthReporter)}
}

func (h *appContextHealth) Aggregate() health.ServiceHealth {
	h.mu.RLock()
	services := make(map[string]HealthReporter, len(h.services))
	for name, reporter := range h.services {
		services[name] = reporter
	}
	h.mu.RUnlock()

	result := health.ServiceHealth{
		Status: health.Up,
	}

	if len(services) == 0 {
		return result
	}
	result.Components = make(map[string]health.ServiceHealth)
	worstRank := 0

	for serviceName, healthReporter := range services {
		serviceHealth := healthReporter.Health()
		rank := 2
		switch serviceHealth.Status {
		case health.Up:
			rank = 0
		case health.Partially:
			rank = 1
		case health.Down:
			rank = 1
		case health.DownCritical:
			rank = 2
		}
		if rank > worstRank {
			worstRank = rank
		}
		result.Components[serviceName] = serviceHealth
	}
	switch worstRank {
	case 0:
		result.Status = health.Up
	case 1:
		result.Status = health.Partially
	default:
		result.Status = health.Down
	}

	return result
}

func (h *appContextHealth) registerHealthReporter(service *reflectiveServiceWrapper) {
	sInstance := service.unwrap()
	if healthReporter, ok := sInstance.(HealthReporter); ok {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.services[service.Name()] = healthReporter
	}
}

func (h *appContextHealth) clear() {
	h.mu.Lock()
	h.services = nil
	h.mu.Unlock()
}
