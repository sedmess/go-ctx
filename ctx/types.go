package ctx

import (
	"context"
	"github.com/sedmess/go-ctx/ctx/health"
	"reflect"
)

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

type InitializableE interface {
	Init(serviceProvider ServiceProvider) error
}

type InitializableContext interface {
	Init(context context.Context, serviceProvider ServiceProvider)
}

type InitializableContextE interface {
	Init(context context.Context, serviceProvider ServiceProvider) error
}

type Constructable interface {
	Init()
}

type ConstructableE interface {
	Init() error
}

type ConstructableContext interface {
	Init(context context.Context)
}

type ConstructableContextE interface {
	Init(context context.Context) error
}

type Named interface {
	Name() string
}

type Disposable interface {
	Dispose()
}

type DisposableE interface {
	Dispose() error
}

//type service interface {
//	InitializableE
//	Named
//	DisposableE
//}

type StartAware interface {
	AfterStart()
}

type StopAware interface {
	BeforeStop()
}

type HealthReporter interface {
	Health() health.ServiceHealth
}
