package ctx

import "reflect"

type servicePackageEntry struct {
	name    string
	service any
}

// ServicePackage retains registration order and duplicate entries until the application
// container validates them. Package construction never silently replaces a service.
type ServicePackage struct {
	entries []servicePackageEntry
}

func (p *ServicePackage) ForEach(fn func(service any, name string)) {
	for _, entry := range p.entries {
		fn(entry.service, entry.name)
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
	pkg := ServicePackage{entries: make([]servicePackageEntry, 0, len(services))}
	for _, svc := range services {
		if named, ok := svc.(NamedService); ok {
			pkg.entries = append(pkg.entries, servicePackageEntry{name: named.name, service: named.svc})
		} else {
			pkg.entries = append(pkg.entries, servicePackageEntry{service: svc})
		}
	}
	return pkg
}
