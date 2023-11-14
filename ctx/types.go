package ctx

import "reflect"

type ServiceProvider interface {
	ByName(name string) any
	ByType(sType any) any
	byReflectType(sType reflect.Type) any
}

type serviceProviderImpl func(string) any

func (p serviceProviderImpl) ByName(name string) any {
	return p(name)
}

func (p serviceProviderImpl) ByType(sType any) any {
	return p(reflect.TypeOf(sType).String())
}

func (p serviceProviderImpl) byReflectType(sType reflect.Type) any {
	typeName := sType.String()
	return p.ByName(typeName)
}

type Initializable interface {
	Init(serviceProvider ServiceProvider)
}

type Named interface {
	Name() string
}

type Disposable interface {
	Dispose()
}

type Service interface {
	Initializable
	Named
	Disposable
}

type LifecycleAware interface {
	AfterStart()
	BeforeStop()
}
