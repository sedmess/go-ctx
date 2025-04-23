package main

import (
	"fmt"
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/health"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u"
	"log/slog"
	"os"
	"time"
)

const aServiceName = "a_service"
const paramAName = "PARAM_A"

type aService struct {
	paramA int
}

func (instance *aService) Init(_ ctx.ServiceProvider) {
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

func (instance *aService) Health() health.ServiceHealth {
	return health.Status(health.Up)
}

const bServiceName = "b_service"

type bService struct {
	a *aService
}

func (instance *bService) Init(serviceProvider ctx.ServiceProvider) {
	instance.a = serviceProvider.ByName(aServiceName).(*aService)
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

func (instance *timedService) Init(_ ctx.ServiceProvider) {
	instance.l = logger.New(instance.Name())

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
		logger.Warn("timer", "onTimer!")
	})
	instance.l.Info("afterStart2")
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

func (instance *appLCService) Init(serviceProvider ctx.ServiceProvider) {
	instance.b = serviceProvider.ByName(bServiceName).(*bService)
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

func (instance *connAService) Init(serviceProvider ctx.ServiceProvider) {
	instance.b = serviceProvider.ByName(bServiceName).(*bService)
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

func (instance *connBService) Init(_ ctx.ServiceProvider) {
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
	name   string
	custom string
}

func (instance *multiInstanceService) Init(_ ctx.ServiceProvider) {
	logger.Info(multiInstanceServiceNamePrefix+instance.name, "init:", ctx.GetEnvCustom(instance.custom, "MIS"))
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

func (instance *multiInstanceGetService) Init(serviceProvider ctx.ServiceProvider) {
	instance.m1 = serviceProvider.ByName(multiInstanceServiceNamePrefix + "1").(*multiInstanceService)
	instance.m2 = serviceProvider.ByName(multiInstanceServiceNamePrefix + "2").(*multiInstanceService)
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

type ReflectiveSingletonService interface {
	Do() string
}

type reflectiveSingletonServiceImpl struct {
	ReflectiveSingletonService
	L logger.Logger `ctx:"singleton"`
	A *aService     `ctx:"a_service"`
}

func (instance *reflectiveSingletonServiceImpl) Name() string {
	return u.GetInterfaceName[ReflectiveSingletonService]()
}

func (instance *reflectiveSingletonServiceImpl) AfterStart() {
	instance.A.Do()
	instance.L.Info("A =", instance.A.Name())
}

func (instance *reflectiveSingletonServiceImpl) BeforeStop() {
	instance.L.Info("stop")
}

func (instance *reflectiveSingletonServiceImpl) Do() string {
	return "done"
}

type reflectiveSingletonService2 struct {
	l logger.Logger              `ctx:""`
	d ReflectiveSingletonService `ctx:""`
}

func (instance *reflectiveSingletonService2) Do() {
	instance.l.Info(instance.d.Do())
}

type panicService struct {
}

func (p *panicService) AfterStart() {
	ctx.Run(func() {
		<-time.After(60 * time.Second)
		panic("for no particular reason")
	})
}

func (p *panicService) BeforeStop() {
}

type anonymousService struct {
	L logger.Logger `ctx:""`
}

func (as *anonymousService) Do(who string) {
	as.L.Info("do for", who)
}

type asConsumerService struct {
	AnonymousService *anonymousService `ctx:""`
	L                logger.Logger     `ctx:""`
}

func (a *asConsumerService) Init(ctx.ServiceProvider) {
	a.L.Info("side actions")
	//a.AnonymousService = serviceProvider.ByType((*anonymousService)(nil)).(*anonymousService)
}

func (a *asConsumerService) AfterStart() {
	a.AnonymousService.Do("asConsumerService")
}

func (a *asConsumerService) BeforeStop() {
}

type loggerDemoService struct {
	l      logger.Logger `ctx:""`
	lNamed logger.Logger `ctx:"named-logger"`
}

func (l *loggerDemoService) AfterStart() {
	l.l.Debug("debug demo", 1)
	l.lNamed.Debug("debug demo", 3)
	logger.Debug("tag-logger", "debug demo", 5)
	l.l.Info("info demo", 1)
	l.lNamed.Info("info demo", 3)
	logger.Info("tag-logger", "info demo", 5)

	l.l.Error("error demo", 1)
	l.lNamed.Error("error demo", 2)
	logger.Error("tag-logger", "error demo", 3)
}

func (l *loggerDemoService) BeforeStop() {
}

type envInjectDemoService struct {
	l                logger.Logger            `ctx:""`
	envValue         *ctx.EnvValue            `env:"DURATION"`
	envValueDuration time.Duration            `env:"DURATION"`
	envValueString   string                   `env:"DURATION"`
	envMap           map[string]*ctx.EnvValue `env:"MAP"`
}

func (e *envInjectDemoService) AfterStart() {
	e.l.Info(e.envValue.AsDuration().String())
}

func (e *envInjectDemoService) BeforeStop() {
}

type ctxInjectService struct {
	l    logger.Logger  `ctx:""`
	ctx1 ctx.AppContext `ctx:"CTX"`
	ctx2 ctx.AppContext
}

func (instance *ctxInjectService) AfterStart() {
	go func() {
		instance.l.Info(instance.ctx1.Stats())
		instance.l.Info(instance.ctx1.Health().Aggregate())
	}()
}

func (instance *ctxInjectService) Init(serviceProvider ctx.ServiceProvider) {
	instance.ctx2 = serviceProvider.ByName("CTX").(ctx.AppContext)
}

type envDefInjectService struct {
	logger.Logger `ctx:""`
	val1          time.Duration            `env:"DEF_VALUE_TEST1=10s"`
	val2          string                   `env:"DEF_VALUE_TEST2=str"`
	val3          map[string]*ctx.EnvValue `env:"DEF_VALUE_TEST3=k1=1,2,3|k2=123|k3=10s"`
	val4          string                   `env:"DURATION=0s"`
	val5          map[string]bool          `env:"STR_SET"`
	val6          map[string]bool          `env:"STR_SET2=s3,s2,s1"`
	val7          map[int]bool             `env:"INT_SET"`
	val8          map[int64]bool           `env:"INT_SET"`
	val9          map[string]bool          `env:"UNDEFINED_SET"`
	val10         map[string]*ctx.EnvValue `env:"UNDEFINED_MAP"`
}

func (e *envDefInjectService) AfterStart() {
	e.Info("val1 =", e.val1.String())
	e.Info("val2 =", e.val2)
	e.Info("val3.k1 =", e.val3["k1"].AsStringArray())
	e.Info("val3.k3 =", e.val3["k2"].AsInt64())
	e.Info("val3.k1 =", e.val3["k3"].AsDuration().String())
	e.Info(fmt.Sprintf("val5 = %v", e.val5))
	e.Info(fmt.Sprintf("val6 = %v", e.val6))
	e.Info(fmt.Sprintf("val7 = %v", e.val7))
	e.Info(fmt.Sprintf("val8 = %v", e.val8))
	e.Info(fmt.Sprintf("val9 = %v", e.val9))
	e.Info(fmt.Sprintf("val10 = %v", e.val10))
}

func (e *envDefInjectService) BeforeStop() {
}

type intRefService interface {
	DoSomething()
}

type intRefServiceImpl struct {
	intRefService `ctx:"impl"`
	l             logger.Logger `ctx:""`
}

func (i *intRefServiceImpl) DoSomething() {
	i.l.Info("do something")
}

type intRef2Service struct {
	srv intRefService `ctx:""`
}

func (i *intRef2Service) AfterStart() {
	i.srv.DoSomething()
}

func (i *intRef2Service) BeforeStop() {
}

type ConstructableService struct {
	l   logger.Logger  `ctx:""`
	ctx ctx.AppContext `ctx:"inject(CTX)"`
}

func (s *ConstructableService) Init() {
	_, stateName := s.ctx.State()
	s.l.Info("initialization with context state:", stateName)
}

type defEnvValue struct {
	val string `env:"UNDEFINED_ENV_VALUE="`
}

func (s *defEnvValue) Init() {
	logger.Info("DEF_VAL", "val =", s.val)
}

func (s *defEnvValue) AfterStart() {
	logger.Info("DEF_VAL", "val =", ctx.GetEnv("UNDEFINED_ENV_VALUE"))
}

func (s *defEnvValue) BeforeStop() {
}

type slowDisposingService struct {
	l logger.Logger `ctx:""`
}

func (s *slowDisposingService) Dispose() {
	s.l.Info("start disposing...")
	<-time.After(time.Second * 10)
	s.l.Info("...disposed")
}

type envCustomService struct {
	key100 int `env:"KEY100"`
	key101 int `env:"KEY101"`
	key103 int `env:"KEY103"`
}

func (s *envCustomService) Init() {
	println("key100 =", s.key100)
	println("key101 =", s.key101)
	println("key103 =", s.key103)
}

type newTags struct {
	l1 logger.Logger `ctx:""`
	l2 logger.Logger `ctx:"named_logger_2"`
	l3 logger.Logger `ctx:"logger"`
	l4 logger.Logger `ctx:"logger(named_logger_4)"`

	envCustomService1 *envCustomService `ctx:""`
	a1                *aService         `ctx:"a_service"`

	envCustomService2 *envCustomService `ctx:"inject"`
	a2                *aService         `ctx:"inject(a_service)"`

	key100 int           `env:"KEY100"`
	key999 time.Duration `env:"KEY999=999s"`
	key0   string        `env:"KEY0="`
}

func (s *newTags) Init() {
	s.l1.Info("init tags", s.key100, s.key999)
	s.l2.Info("init tags", s.key100, s.key999)
	s.l3.Info("init tags", s.key100, s.key999)
	s.l4.Info("init tags", s.key100, s.key999)
	s.l1.Info(s.envCustomService1 == s.envCustomService2)
	s.l1.Info(s.a1 == s.a2)
	s.l1.Info("key0 =", s.key0)
}

type slogExample struct {
	l1 *slog.Logger  `ctx:""`
	l2 *slog.Logger  `ctx:"named_slog1"`
	l3 *slog.Logger  `ctx:"logger"`
	l4 *slog.Logger  `ctx:"logger(named_slog2)"`
	l5 *slog.Logger  `ctx:"logger(named_slog3) loggerAttr(tag1=val1) loggerAttr(tag2=val2) loggerAttr(tag3=val3)"`
	l6 logger.Logger `ctx:"logger loggerAttr(tag1=val1) loggerAttr(tag2=val2) loggerAttr(tag3=val3)"`
}

func (s *slogExample) Init() {
	s.l1.Info("test", slog.String("test", "hello world"))
	s.l2.Info("test")
	s.l3.Info("test")
	s.l4.Info("test")
	s.l5.Info("test")
	s.l6.Warn("test", 6)
}

type envStruct struct {
	l      logger.Logger `ctx:""`
	param1 string        `env:"PARAM1=param1"`
	key100 int           `env:"KEY100"`
}

func main() {
	_ = os.Setenv("SLOG_LEVEL", "debug")
	_ = os.Setenv("SLOG_ADD_SOURCE", "true")
	_ = os.Setenv("SLOG_ADD_COMMON_TAGS", "true")
	_ = os.Setenv("SLOG_HANDLER", "legacy")

	ctx.SetSlogWriter(
		os.Stdout,
		os.Stderr,
	)

	_ = os.Setenv("MAP", "key1=value1|key2=123")
	envMap := ctx.GetEnv("map").AsMap()
	println(envMap["key1"].AsString())
	println(envMap["key2"].AsInt())

	_ = os.Setenv("MIS", "mis_default")
	_ = os.Setenv("I1_MIS", "mis_i1")
	_ = os.Setenv("DURATION", "60s")
	_ = os.Setenv("STR_SET", "s1,s2,s3")
	_ = os.Setenv("INT_SET", "1,2,3")

	println(ctx.GetEnv("DURATION").AsDuration().String())
	println(fmt.Sprintf("%v", ctx.GetEnv("STR_SET").AsStringSet()))
	println(fmt.Sprintf("%v", ctx.GetEnv("STR_SET").AsStringSet()))
	println(fmt.Sprintf("%v", ctx.GetEnv("INT_SET").AsIntSet()))
	println(fmt.Sprintf("%v", ctx.GetEnv("INT_SET").AsInt64Set()))

	connAService := newConnAService()
	go func() {
		time.Sleep(5 * time.Second)
		connAService.Send("a")
	}()

	r2 := &reflectiveSingletonService2{}

	go func() {
		time.Sleep(10 * time.Second)
		r2.Do()
	}()

	go func() {
		<-time.After(5 * time.Second)
		aService := ctx.GetService(aServiceName).(*aService)
		aService.Do()
	}()

	es := ctx.Env[envStruct]()
	println(es.param1)
	println(es.key100)
	println(es.l)

	application := ctx.CreateContextualizedApplication(
		ctx.PackageOf(
			&aService{}, &bService{}, &timedService{}, &appLCService{}, connAService, newConnBService(), &multiInstanceService{name: multiInstanceServiceNamePrefix + "1", custom: "I1"}, &multiInstanceService{name: multiInstanceServiceNamePrefix + "2", custom: "I2"}, &multiInstanceGetService{},
			&reflectiveSingletonServiceImpl{}, r2,
			&panicService{},
			&anonymousService{}, &asConsumerService{},
			&loggerDemoService{},
			&envInjectDemoService{},
			&ctxInjectService{},
			&envDefInjectService{},
			&intRefServiceImpl{},
			&intRef2Service{},
			&ConstructableService{},
			&defEnvValue{},
			&slowDisposingService{},
			&envCustomService{},
			&newTags{},
			&slogExample{},
		),
		ctx.PackageOf(ctx.ConnectServices(connAServiceName, connBServiceName)),
	)

	go func() {
		application.Join()
		println("application stopped 2")
	}()

	go func() {
		<-time.After(time.Second * 60)
		application.Stop()
		application.Join()
		println("application stopped 3")
		application.Join()
	}()

	application.Join()

	println("application stopped")
}
