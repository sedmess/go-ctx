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

func main() {
	ctx.StartContextualizedApplication([]ctx.Service{&aService{}, &bService{}, &timedService{}, &appLCService{}})
}
