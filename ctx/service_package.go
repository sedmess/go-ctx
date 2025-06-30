package ctx

import "reflect"

type ServicePackage struct {
	namedServices map[string]any
	services      []any
}

func (p *ServicePackage) ForEach(fn func(service any, name string)) {
	for name, service := range p.namedServices {
		fn(service, name)
	}
	for _, service := range p.services {
		fn(service, "")
	}
}

type NamedService struct {
	name string
	svc  any
}

func (n *NamedService) Name() string {
	return n.name
}

func (n *NamedService) Service() any {
	return n.svc
}

func WithName(name string, service any) NamedService {
	return NamedService{name: name, svc: service}
}

func Typed[T any]() NamedService {
	typeFor := reflect.TypeFor[T]()
	value := reflect.New(typeFor).Interface()
	return NamedService{svc: value}
}

func TypedWithName[T any](name string) NamedService {
	s := Typed[T]()
	s.name = name
	return s
}

func PackageOf(services ...any) ServicePackage {
	pkg := ServicePackage{
		namedServices: make(map[string]any),
		services:      make([]any, 0),
	}

	for _, svc := range services {
		if named, ok := svc.(NamedService); ok {
			if named.name != "" {
				pkg.namedServices[named.name] = named.svc
			} else {
				pkg.services = append(pkg.services, named.svc)
			}
		} else {
			pkg.services = append(pkg.services, svc)
		}
	}

	return pkg
}
