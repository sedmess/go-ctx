package actuator

import (
	"github.com/sedmess/go-ctx/base/httpserver"
	"github.com/sedmess/go-ctx/ctx"
	"reflect"
	"sync"
)

const defaultServerName = "base.actuator-http-server"

func AddToDefaultHttpServer() any {
	return &controller{serverServiceName: reflect.TypeOf(new(httpserver.RestServer)).Elem().String()}
}

func AddToHttpServer(serverServiceName string) any {
	return &controller{serverServiceName: serverServiceName}
}

var independentServerServices = sync.OnceValue(func() ctx.ServicePackage {
	return ctx.PackageOf(
		httpserver.NewRestServer(defaultServerName, "ACTUATOR"),
		&controller{serverServiceName: defaultServerName},
	)
})

func RunAsIndependentServer() ctx.ServicePackage {
	return independentServerServices()
}
