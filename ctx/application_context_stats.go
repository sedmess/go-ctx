package ctx

import (
	"fmt"
	"reflect"
	"sync"
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

func createDescriptorFor(service *reflectiveServiceWrapper) ServiceDescriptor {
	sInstance := service.unwrap()
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
	for _, dependency := range sd.Dependencies {
		if dependency == serviceName {
			return
		}
	}
	sd.Dependencies = append(sd.Dependencies, serviceName)
}

type appContextStats struct {
	mu       sync.RWMutex
	services map[string]ServiceDescriptor
}

func createContextStats() *appContextStats {
	return &appContextStats{services: make(map[string]ServiceDescriptor)}
}

func (s *appContextStats) Services() map[string]ServiceDescriptor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyServiceDescriptors(s.services)
}

func (s *appContextStats) String() string {
	return fmt.Sprintf("services: %v", s.Services())
}

func (s *appContextStats) registerServiceDescriptor(sd ServiceDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sd.Dependencies = append([]string(nil), sd.Dependencies...)
	s.services[sd.Name] = sd
}

func (s *appContextStats) dependencySnapshot() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string][]string, len(s.services))
	for name, descriptor := range s.services {
		result[name] = append([]string(nil), descriptor.Dependencies...)
	}
	return result
}

func copyServiceDescriptors(source map[string]ServiceDescriptor) map[string]ServiceDescriptor {
	result := make(map[string]ServiceDescriptor, len(source))
	for name, descriptor := range source {
		descriptor.Dependencies = append([]string(nil), descriptor.Dependencies...)
		result[name] = descriptor
	}
	return result
}
