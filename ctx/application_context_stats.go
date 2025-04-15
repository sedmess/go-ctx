package ctx

import (
	"fmt"
	"reflect"
)

type AppContextStats interface {
	Services() map[string]ServiceDescriptor
}

type ServiceDescriptor struct {
	Name         string
	Type         reflect.Type
	IsStartAware bool
	IsStopAware  bool
	Dependencies []string
}

func createDescriptorFor(service Service) ServiceDescriptor {
	sInstance := unwrap(service)
	_, isStartAware := sInstance.(StartAware)
	_, isStopAware := sInstance.(StopAware)
	return ServiceDescriptor{
		Name:         service.Name(),
		Type:         reflect.TypeOf(sInstance),
		IsStartAware: isStartAware,
		IsStopAware:  isStopAware,
		Dependencies: make([]string, 0),
	}
}

func (sd *ServiceDescriptor) addDependency(serviceName string) {
	sd.Dependencies = append(sd.Dependencies, serviceName)
}

type appContextStats struct {
	services map[string]ServiceDescriptor
}

func createContextStats() *appContextStats {
	return &appContextStats{services: make(map[string]ServiceDescriptor)}
}

func (s *appContextStats) Services() map[string]ServiceDescriptor {
	return s.services
}

func (s *appContextStats) String() string {
	return fmt.Sprintf("services: %v", s.services)
}

func (s *appContextStats) registerServiceDescriptor(sd ServiceDescriptor) {
	s.services[sd.Name] = sd
}
