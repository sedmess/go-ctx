package ctx

type ServiceProvider func(serviceName string) any

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
