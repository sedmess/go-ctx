package ctx

import (
	"strconv"
	"time"
)

type Connectable interface {
	LifecycleAware
	Ingoing(inChan chan interface{})
	Outgoing() chan interface{}
	OnMessage(msg interface{})
	Send(msg interface{})
}

type BasicConnector struct {
	inCh  chan interface{}
	outCh chan interface{}
	qCh   chan bool
}

func (connector *BasicConnector) Init(_ func(string) Service) {
	connector.outCh = make(chan interface{})
}

func (connector *BasicConnector) Ingoing(inChan chan interface{}) {
	connector.inCh = inChan
}

func (connector *BasicConnector) OnMessage(msg interface{}) {
	panic("not implemented")
}

func (connector *BasicConnector) AfterStart() {
	connector.listen(connector.OnMessage)
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
				onMessage(msg)
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

func ConnectServices(services ...string) []Service {
	if len(services)%2 == 1 {
		panic("wrong arguments")
	}
	pairs := make([][]string, 0)
	for i := 0; i < len(services); i += 2 {
		pairs = append(pairs, []string{services[i], services[i+1]})
	}
	return []Service{
		&mutualConnectableConnector{
			name:  mutualConnectableConnectorNamePrefix + strconv.FormatInt(time.Now().Unix(), 36),
			pairs: pairs,
		},
	}
}

func (instance *mutualConnectableConnector) Init(serviceProvider func(serviceName string) Service) {
	for _, pair := range instance.pairs {
		service1, ok := serviceProvider(pair[0]).(Connectable)
		if !ok {
			panic(pair[0] + " not Connectable")
		}
		service2, ok := serviceProvider(pair[1]).(Connectable)
		if !ok {
			panic(pair[1] + " not Connectable")
		}
		service1.Ingoing(service2.Outgoing())
		service2.Ingoing(service1.Outgoing())
		LogInfo(instance.name, "connected services:", pair[0], pair[1])
	}
}

func (instance *mutualConnectableConnector) Name() string {
	return instance.name
}

func (instance *mutualConnectableConnector) Dispose() {
}
