package health

import "fmt"

type ServiceHealthStatus string

const (
	Up           = ServiceHealthStatus("UP")
	Partially    = ServiceHealthStatus("PARTIALLY")
	DownCritical = ServiceHealthStatus("DOWN_CRITICAL")
	Down         = ServiceHealthStatus("DOWN")
)

type ServiceHealth struct {
	Status     ServiceHealthStatus      `json:"status,omitempty"`
	Details    map[string]any           `json:"details,omitempty"`
	Components map[string]ServiceHealth `json:"components,omitempty"`
}

func Status(status ServiceHealthStatus) ServiceHealth {
	return ServiceHealth{Status: status}
}

func (s ServiceHealth) String() string {
	return fmt.Sprintf("health: %v Details: %v Components: %v", s.Status, s.Details, s.Components)
}
