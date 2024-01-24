package actuator

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sedmess/go-ctx/base/httpserver"
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/logger"
	"net/http"
	"sync/atomic"
)

const controllerName = "base.actuator-controller"

type controller struct {
	l logger.Logger `logger:""`

	serverServiceName string

	appContext ctx.AppContext `inject:"CTX"`

	ctxRunning atomic.Bool
}

func (instance *controller) Init(provider ctx.ServiceProvider) {
	server := provider.ByName(instance.serverServiceName).(httpserver.RestServer)

	instance.l.Info("register on server", instance.serverServiceName)

	server.RegisterRoute(rest.Get("/actuator/health", httpserver.RequestHandler[any](instance.l).Handle(instance.health)))
	server.RegisterRoute(rest.Get("/actuator/services", httpserver.RequestHandler[any](instance.l).Handle(instance.services)))

	instance.ctxRunning.Store(false)
}

func (instance *controller) Name() string {
	return controllerName
}

func (instance *controller) AfterStart() {
	instance.ctxRunning.Store(true)
}

func (instance *controller) BeforeStop() {
	instance.ctxRunning.Store(false)
}

func (instance *controller) health() (any, int, error) {
	return instance.appContext.Health().Aggregate(), http.StatusOK, nil
}

func (instance *controller) services() (any, int, error) {
	services := instance.appContext.Stats().Services()
	result := make(map[string]any)
	for _, descriptor := range services {
		srv := make(map[string]any)
		srv["name"] = descriptor.Name
		srv["type"] = descriptor.Type.String()
		srv["isLifecycleAware"] = descriptor.IsLifecycleAware
		if len(descriptor.Dependencies) > 0 {
			srv["dependencies"] = descriptor.Dependencies
		}
		result[descriptor.Name] = srv
	}
	return result, http.StatusOK, nil
}
