package ctx

import (
	"github.com/sedmess/go-ctx/logger"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type connectable interface {
	ingoing(inChan chan any)
	outgoing() chan any
	inType() reflect.Type
	outType() reflect.Type
}

type ConnectableService[In any, Out any] interface {
	Service
	OnMessage(In)
	Send(Out)
}

type ServiceConnector[In any, Out any] struct {
	name              string
	inCh              chan any
	outCh             chan any
	qCh               chan bool
	onMessageListener func(msg In)
}

func NewServiceConnector[In any, Out any](service ConnectableService[In, Out]) ServiceConnector[In, Out] {
	return ServiceConnector[In, Out]{name: service.Name(), onMessageListener: service.OnMessage, outCh: make(chan any)}
}

func (connector *ServiceConnector[In, Out]) AfterStart() {
	connector.listen(connector.onMessageListener)
}

func (connector *ServiceConnector[In, Out]) BeforeStop() {
	connector.stopListening()
}

func (connector *ServiceConnector[In, Out]) Send(msg Out) {
	connector.outCh <- msg
}

func (connector *ServiceConnector[In, Out]) ingoing(inChan chan any) {
	connector.inCh = inChan
}

func (connector *ServiceConnector[In, Out]) outgoing() chan any {
	return connector.outCh
}

func (connector *ServiceConnector[In, Out]) inType() reflect.Type {
	var in In
	return reflect.TypeOf(in)
}

func (connector *ServiceConnector[In, Out]) outType() reflect.Type {
	var out Out
	return reflect.TypeOf(out)
}

func (connector *ServiceConnector[In, Out]) listen(onMessage func(msg In)) {
	connector.qCh = make(chan bool)
	go func() {
		for {
			select {
			case msg := <-connector.inCh:
				runWithRecover(
					func() {
						onMessage(msg.(In))
					},
					func(reason any) {
						logger.Error(connector.name, "during onMessage:", reason)
						panic(reason)
					},
				)
			case <-connector.qCh:
				break
			}
		}
	}()
}

func (connector *ServiceConnector[In, Out]) stopListening() {
	connector.qCh <- true
}

const mutualConnectableConnectorNamePrefix = "_connector_"

type mutualConnectableConnector struct {
	name  string
	pairs [][]string
}

func ConnectServices(services ...string) any {
	if len(services)%2 == 1 {
		panic("wrong arguments")
	}
	pairs := make([][]string, 0)
	for i := 0; i < len(services); i += 2 {
		pairs = append(pairs, []string{services[i], services[i+1]})
	}
	return &mutualConnectableConnector{
		name:  mutualConnectableConnectorNamePrefix + strconv.FormatInt(time.Now().UnixNano(), 36),
		pairs: pairs,
	}
}

func (instance *mutualConnectableConnector) Init(serviceProvider ServiceProvider) {
	for _, pair := range instance.pairs {
		service1, ok := serviceProvider(pair[0]).(connectable)
		if !ok {
			panic(pair[0] + " can't be connected")
		}
		service2, ok := serviceProvider(pair[1]).(connectable)
		if !ok {
			panic(pair[1] + " can't be connected")
		}

		srvsStr := strings.Join([]string{pair[0], pair[1]}, ",")

		if service1.inType() != service2.outType() || service1.outType() != service2.inType() {
			panic(srvsStr + " can't be connected")
		}

		service1.ingoing(service2.outgoing())
		service2.ingoing(service1.outgoing())
		logger.Debug(instance.name, "connected services:", srvsStr)
	}
}

func (instance *mutualConnectableConnector) Name() string {
	return instance.name
}

func (instance *mutualConnectableConnector) Dispose() {
}
