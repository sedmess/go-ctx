package ctx

import "github.com/sedmess/go-ctx/ctx/health"

type AppContextHealth interface {
	Aggregate() health.ServiceHealth
}

type appContextHealth struct {
	services map[string]HealthReporter
}

func createContextHealth() *appContextHealth {
	return &appContextHealth{services: make(map[string]HealthReporter)}
}

func (h *appContextHealth) Aggregate() health.ServiceHealth {
	result := health.ServiceHealth{
		Status: health.Up,
	}

	if len(h.services) == 0 {
		return result
	}
	result.Components = make(map[string]health.ServiceHealth)

	for serviceName, healthReporter := range h.services {
		serviceHealth := healthReporter.Health()
		switch serviceHealth.Status {
		case health.Partially:
			result.Status = health.Partially
		case health.Down:
			result.Status = health.Partially
		case health.DownCritical:
			result.Status = health.Down
		}
		result.Components[serviceName] = serviceHealth
	}

	return result
}

func (h *appContextHealth) registerHealthReporter(name string, service HealthReporter) {
	h.services[name] = service
}
