package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/logger"
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
	logger.Info(instance.Name(), "initialized")
}

func (instance *aService) Name() string {
	return aServiceName
}

func (instance *aService) Dispose() {
	logger.Info(instance.Name(), "disposed")
}

func (instance *aService) Do() {
	logger.Info(instance.Name(), "invoked: paramA =", instance.paramA)
}

const bServiceName = "b_service"

type bService struct {
	a *aService
}

func (instance *bService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.a = serviceProvider(aServiceName).(*aService)
	logger.Info(instance.Name(), "initialized")
}

func (instance *bService) Name() string {
	return bServiceName
}

func (instance *bService) Dispose() {
	logger.Info(instance.Name(), "disposed")
}

func (instance *bService) Do() {
	logger.Info(instance.Name(), "invoked")
	instance.a.Do()
}

const timedServiceName = "timed_service"

type timedService struct {
	ctx.TimerTask

	l logger.Logger
}

func (instance *timedService) Init(_ func(serviceName string) ctx.Service) {
	instance.l = logger.New(instance)

	instance.l.Info("initialized")
}

func (instance *timedService) Name() string {
	return timedServiceName
}

func (instance *timedService) Dispose() {
	instance.l.Info("disposed")
}

func (instance *timedService) AfterStart() {
	instance.l.Info("afterStart")
	instance.StartTimer(2*time.Second, func() {
		logger.Info("timer", "onTimer!")
	})
}

func (instance *timedService) BeforeStop() {
	instance.l.Info("beforeStop")
	time.Sleep(3 * time.Second)
	instance.StopTimer()
}

const appLCServiceName = "app_lc_service"

type appLCService struct {
	b *bService
}

func (instance *appLCService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.b = serviceProvider(bServiceName).(*bService)
	logger.Info(instance.Name(), "initialized")
}

func (instance *appLCService) Name() string {
	return appLCServiceName
}

func (instance *appLCService) Dispose() {
	logger.Info(instance.Name(), "disposed")
}

func (instance *appLCService) AfterStart() {
	logger.Info(appLCServiceName, "app started")
	instance.b.Do()
}

func (instance *appLCService) BeforeStop() {
	logger.Info(appLCServiceName, "app stopped")
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
	logger.Info(connAServiceName, "msg: "+msg)
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
	logger.Info(connBServiceName, "msg: "+msg)
	instance.Send(msg + "b")
}

const multiInstanceServiceNamePrefix = "multi_instance_service_"

type multiInstanceService struct {
	name string
}

func (instance *multiInstanceService) Init(func(serviceName string) ctx.Service) {
	logger.Info(multiInstanceServiceNamePrefix+instance.name, "init")
}

func (instance *multiInstanceService) Name() string {
	return instance.name
}

func (instance *multiInstanceService) Dispose() {
	logger.Info(multiInstanceServiceNamePrefix+instance.name, "dispose")
}

const multiInstanceGetServiceName = "multi_instance_get_service"

type multiInstanceGetService struct {
	m1 *multiInstanceService
	m2 *multiInstanceService
}

func (instance *multiInstanceGetService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.m1 = serviceProvider(multiInstanceServiceNamePrefix + "1").(*multiInstanceService)
	instance.m2 = serviceProvider(multiInstanceServiceNamePrefix + "2").(*multiInstanceService)
}

func (instance *multiInstanceGetService) Name() string {
	return multiInstanceGetServiceName
}

func (instance *multiInstanceGetService) AfterStart() {
	logger.Info(multiInstanceGetServiceName, "deps m1:", instance.m1.Name())
	logger.Info(multiInstanceGetServiceName, "deps m2:", instance.m2.Name())
}

func (instance *multiInstanceGetService) BeforeStop() {
}

func (instance *multiInstanceGetService) Dispose() {
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
			&aService{}, &bService{}, &timedService{}, &appLCService{}, connAService, newConnBService(), &multiInstanceService{multiInstanceServiceNamePrefix + "1"}, &multiInstanceService{multiInstanceServiceNamePrefix + "2"}, &multiInstanceGetService{},
		},
		ctx.ServiceArray(ctx.ConnectServices(connAServiceName, connBServiceName)),
	)
}
