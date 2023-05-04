package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"os"
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
	ctx.LogInfo(instance.Name(), "initialized")
}

func (instance *timedService) Name() string {
	return timedServiceName
}

func (instance *timedService) Dispose() {
	ctx.LogInfo(instance.Name(), "disposed")
}

func (instance *timedService) AfterStart() {
	ctx.LogInfo(timedServiceName, "afterStart")
	instance.StartTimer(2*time.Second, func() {
		ctx.LogInfo("timer", "onTimer!")
	})
}

func (instance *timedService) BeforeStop() {
	ctx.LogInfo(timedServiceName, "beforeStop")
	time.Sleep(3 * time.Second)
	instance.StopTimer()
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
	ctx.LogInfo(appLCServiceName, "app started")
	instance.b.Do()
}

func (instance *appLCService) BeforeStop() {
	ctx.LogInfo(appLCServiceName, "app stopped")
}

const connAServiceName = "conn_a_service"

type connAService struct {
	ctx.ServiceConnector[string, string]
	b *bService
}

func newConnAService() *connAService {
	service := &connAService{}
	service.ServiceConnector = ctx.NewServiceConnector[string, string](service)
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

func (instance *connAService) OnMessage(msg string) {
	ctx.LogInfo(connAServiceName, "msg: "+msg)
	instance.b.Do()
}

const connBServiceName = "conn_b_service"

type connBService struct {
	ctx.ServiceConnector[string, string]
}

func newConnBService() *connBService {
	service := &connBService{}
	service.ServiceConnector = ctx.NewServiceConnector[string, string](service)
	return service
}

func (instance *connBService) Init(_ func(msg string) ctx.Service) {
}

func (instance *connBService) Name() string {
	return connBServiceName
}

func (instance *connBService) Dispose() {
}

func (instance *connBService) OnMessage(msg string) {
	ctx.LogInfo(connBServiceName, "msg: "+msg)
	instance.Send(msg + "b")
}

func main() {
	_ = os.Setenv("map", "key1=value1|key2=123")
	envMap := ctx.GetEnv("map").AsMap()
	println(envMap["key1"].AsString())
	println(envMap["key2"].AsInt())

	connAService := newConnAService()
	go func() {
		time.Sleep(5 * time.Second)
		connAService.Send("a")
	}()

	ctx.StartContextualizedApplication(
		[]ctx.Service{
			&aService{}, &bService{}, &timedService{}, &appLCService{}, connAService, newConnBService(),
		},
		ctx.ServiceArray(ctx.ConnectServices(connAServiceName, connBServiceName)),
	)
}
