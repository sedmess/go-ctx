package ctx

import (
	"strconv"
	"time"
)

type ConnectableService interface {
	LifecycleAware
	Ingoing(inChan chan interface{})
	Outgoing() chan interface{}
	Send(msg interface{})
}

type BasicConnector struct {
	name              string
	inCh              chan interface{}
	outCh             chan interface{}
	qCh               chan bool
	onMessageListener func(msg interface{})
}

func NewBasicConnector(name string, listener func(msg interface{})) BasicConnector {
	return BasicConnector{name: name, onMessageListener: listener, outCh: make(chan interface{})}
}

func (connector *BasicConnector) Ingoing(inChan chan interface{}) {
	connector.inCh = inChan
}

func (connector *BasicConnector) AfterStart() {
	connector.listen(connector.onMessageListener)
}

func (connector *BasicConnector) BeforeStop() {
	connector.stopListening()
}

func (connector *BasicConnector) Send(msg interface{}) {
	connector.outCh <- msg
}

func (connector *BasicConnector) Outgoing() chan interface{} {
	return connector.outCh
}

func (connector *BasicConnector) listen(onMessage func(msg interface{})) {
	connector.qCh = make(chan bool)
	go func() {
		for {
			select {
			case msg := <-connector.inCh:
				runWithRecover(
					func() {
						onMessage(msg)
					},
					func(err error) {
						LogError(connector.name, "during onMessage:", err)
					},
				)
			case <-connector.qCh:
				break
			}
		}
	}()
}

func (connector *BasicConnector) stopListening() {
	connector.qCh <- true
}

const mutualConnectableConnectorNamePrefix = "_connector_"

type mutualConnectableConnector struct {
	name  string
	pairs [][]string
}

func ConnectServices(services ...string) Service {
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

func (instance *mutualConnectableConnector) Init(serviceProvider func(serviceName string) Service) {
	for _, pair := range instance.pairs {
		service1, ok := serviceProvider(pair[0]).(ConnectableService)
		if !ok {
			panic(pair[0] + " not ConnectableService")
		}
		service2, ok := serviceProvider(pair[1]).(ConnectableService)
		if !ok {
			panic(pair[1] + " not ConnectableService")
		}
		service1.Ingoing(service2.Outgoing())
		service2.Ingoing(service1.Outgoing())
		LogDebug(instance.name, "connected services:", pair[0], pair[1])
	}
}

func (instance *mutualConnectableConnector) Name() string {
	return instance.name
}

func (instance *mutualConnectableConnector) Dispose() {
}
