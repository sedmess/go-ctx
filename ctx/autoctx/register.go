package autoctx

import (
	"slices"
	"sync"
)

var services []any
var servicesMu sync.Mutex

func S(service any) {
	servicesMu.Lock()
	services = append(services, service)
	servicesMu.Unlock()
}

func RegisteredServices() (res []any) {
	servicesMu.Lock()
	res = slices.Clone(services)
	servicesMu.Unlock()
	return
}
