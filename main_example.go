package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"time"
)

const aServiceName = "a_service"

type aService struct {
}

func (instance *aService) Init(_ func(serviceName string) ctx.Service) {
	println(instance.Name(), "initialized")
}

func (instance *aService) Name() string {
	return aServiceName
}

func (instance *aService) Dispose() {
	println(instance.Name(), "disposed")
}

func (instance *aService) Do() {
	println(instance.Name(), "invoked")
}

const bServiceName = "b_service"

type bService struct {
	a *aService
}

func (instance *bService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.a = serviceProvider(aServiceName).(*aService)
	println(instance.Name(), "initialized")
}

func (instance *bService) Name() string {
	return bServiceName
}

func (instance *bService) Dispose() {
	println(instance.Name(), "disposed")
}

func (instance *bService) Do() {
	println(instance.Name(), "invoked")
	instance.a.Do()
}

const timedServiceName = "timed_service"

type timedService struct {
	ctx.TimerTask
}

func (instance *timedService) Init(_ func(serviceName string) ctx.Service) {
	instance.StartTimer(2*time.Second, func() {
		println("onTimer!")
	})
	println(instance.Name(), "initialized")
}

func (instance *timedService) Name() string {
	return timedServiceName
}

func (instance *timedService) Dispose() {
	instance.StopTimer()
	println(instance.Name(), "disposed")
}

const appLCSerivceName = "app_lc_service"

type appLCService struct {
	b *bService
}

func (instance *appLCService) Init(serviceProvider func(serviceName string) ctx.Service) {
	instance.b = serviceProvider(bServiceName).(*bService)
	println(instance.Name(), "initialized")
}

func (instance *appLCService) Name() string {
	return appLCSerivceName
}

func (instance *appLCService) Dispose() {
	println(instance.Name(), "disposed")
}

func (instance *appLCService) AfterStart() {
	println("app started")
	instance.b.Do()
}

func (instance *appLCService) BeforeStop() {
	println("app stopped")
}

func main() {
	ctx.StartContextualizedApplication([]ctx.Service{&aService{}, &bService{}, &timedService{}, &appLCService{}})
}
