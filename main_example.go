package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"time"
)

const aServiceName = "a_service"
const paramAName = "PARAM_A"

type aService struct {
	paramA int
}

func (instance *aService) Init(_ func(serviceName string) ctx.Service) {
	instance.paramA = ctx.GetEnv(paramAName).AsIntDefault(5)
	ctx.LogInfo(instance.Name(), "initialized")
}

func (instance *aService) Name() string {
	return aServiceName
}

func (instance *aService) Dispose() {
	ctx.LogInfo(instance.Name(), "disposed")
}

func (instance *aService) Do() {
	ctx.LogInfo(instance.Name(), "invoked: paramA =", instance.paramA)
}

const bServiceName = "b_service"

type bService struct {
	a *aService
}

func (instance *bService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.a = serviceProvider(aServiceName).(*aService)
	ctx.LogInfo(instance.Name(), "initialized")
}

func (instance *bService) Name() string {
	return bServiceName
}

func (instance *bService) Dispose() {
	ctx.LogInfo(instance.Name(), "disposed")
}

func (instance *bService) Do() {
	ctx.LogInfo(instance.Name(), "invoked")
	instance.a.Do()
}

const timedServiceName = "timed_service"

type timedService struct {
	ctx.TimerTask
}

func (instance *timedService) Init(_ func(serviceName string) ctx.Service) {
	instance.StartTimer(2*time.Second, func() {
		ctx.LogInfo("timer", "onTimer!")
	})
	ctx.LogInfo(instance.Name(), "initialized")
}

func (instance *timedService) Name() string {
	return timedServiceName
}

func (instance *timedService) Dispose() {
	instance.StopTimer()
	ctx.LogInfo(instance.Name(), "disposed")
}

const appLCServiceName = "app_lc_service"

type appLCService struct {
	b *bService
}

func (instance *appLCService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.b = serviceProvider(bServiceName).(*bService)
	ctx.LogInfo(instance.Name(), "initialized")
}

func (instance *appLCService) Name() string {
	return appLCServiceName
}

func (instance *appLCService) Dispose() {
	ctx.LogInfo(instance.Name(), "disposed")
}

func (instance *appLCService) AfterStart() {
	ctx.LogInfo("app started")
	instance.b.Do()
}

func (instance *appLCService) BeforeStop() {
	ctx.LogInfo("app stopped")
}

const connAServiceName = "conn_a_service"

type connAService struct {
	ctx.BasicConnector
	b *bService
}

func NewConnAService() *connAService {
	service := &connAService{}
	service.BasicConnector = ctx.NewBasicConnector(service.Name(), service.OnMessage)
	return service
}

func (instance *connAService) Init(serviceProvider func(msg string) ctx.Service) {
	instance.b = serviceProvider(bServiceName).(*bService)
}

func (instance *connAService) Name() string {
	return connAServiceName
}

func (instance *connAService) Dispose() {
}

func (instance *connAService) OnMessage(msg interface{}) {
	str := msg.(string)
	ctx.LogInfo(connAServiceName, "msg: "+str)
	instance.b.Do()
}

const connBServiceName = "conn_b_service"

type connBService struct {
	ctx.BasicConnector
}

func NewConnBService() *connBService {
	service := &connBService{}
	service.BasicConnector = ctx.NewBasicConnector(service.Name(), func(msg interface{}) {
		str := msg.(string)
		ctx.LogInfo(connBServiceName, "msg: "+str)
		service.Send(str + "b")
	})
	return service
}

func (instance *connBService) Init(_ func(msg string) ctx.Service) {
}

func (instance *connBService) Name() string {
	return connBServiceName
}

func (instance *connBService) Dispose() {
}

func main() {
	connAService := NewConnAService()
	go func() {
		time.Sleep(5 * time.Second)
		connAService.Send("a")
	}()

	ctx.StartContextualizedApplication(
		[]ctx.Service{
			&aService{}, &bService{}, &timedService{}, &appLCService{}, connAService, NewConnBService(),
		},
		ctx.ServiceArray(ctx.ConnectServices(connAServiceName, connBServiceName)),
	)
}
