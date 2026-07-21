package ctx

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
)

type connectable interface {
	prepareConnection() <-chan struct{}
	connect(inChan, outChan chan any, peerDone <-chan struct{})
	inType() reflect.Type
	outType() reflect.Type
}

type ConnectableService[In any, Out any] interface {
	Initializable
	Disposable
	Named
	OnMessage(In)
	Send(Out)
}

type connectorGeneration struct {
	stop     chan struct{}
	stopOnce sync.Once
}

func newConnectorGeneration() *connectorGeneration {
	return &connectorGeneration{stop: make(chan struct{})}
}

func (generation *connectorGeneration) close() {
	if generation != nil {
		generation.stopOnce.Do(func() { close(generation.stop) })
	}
}

type ServiceConnector[In any, Out any] struct {
	opMu sync.Mutex
	mu   sync.RWMutex

	name              string
	inCh              chan any
	outCh             chan any
	peerDone          <-chan struct{}
	generation        *connectorGeneration
	listenerDone      chan struct{}
	onMessageListener func(msg In)
}

func NewServiceConnector[In any, Out any](service ConnectableService[In, Out]) ServiceConnector[In, Out] {
	return ServiceConnector[In, Out]{name: service.Name(), onMessageListener: service.OnMessage}
}

func (connector *ServiceConnector[In, Out]) AfterStart() {
	connector.listen(connector.onMessageListener)
}

func (connector *ServiceConnector[In, Out]) BeforeStop() {
	connector.stopListening()
}

// Send preserves channel backpressure while the connection is live. It returns without
// delivery when this endpoint or its peer has begun termination.
func (connector *ServiceConnector[In, Out]) Send(msg Out) {
	connector.mu.RLock()
	outCh := connector.outCh
	peerDone := connector.peerDone
	generation := connector.generation
	connector.mu.RUnlock()
	if outCh == nil || generation == nil {
		return
	}
	select {
	case outCh <- msg:
	case <-generation.stop:
	case <-peerDone:
	}
}

func (connector *ServiceConnector[In, Out]) prepareConnection() <-chan struct{} {
	connector.opMu.Lock()
	defer connector.opMu.Unlock()
	connector.stopAndJoinCurrent()

	generation := newConnectorGeneration()
	connector.mu.Lock()
	connector.generation = generation
	connector.inCh = nil
	connector.outCh = nil
	connector.peerDone = nil
	connector.listenerDone = nil
	connector.mu.Unlock()
	return generation.stop
}

func (connector *ServiceConnector[In, Out]) connect(inChan, outChan chan any, peerDone <-chan struct{}) {
	connector.mu.Lock()
	connector.inCh = inChan
	connector.outCh = outChan
	connector.peerDone = peerDone
	connector.mu.Unlock()
}

func (connector *ServiceConnector[In, Out]) inType() reflect.Type {
	return reflect.TypeFor[In]()
}

func (connector *ServiceConnector[In, Out]) outType() reflect.Type {
	return reflect.TypeFor[Out]()
}

func (connector *ServiceConnector[In, Out]) listen(onMessage func(msg In)) {
	connector.opMu.Lock()
	defer connector.opMu.Unlock()

	connector.mu.Lock()
	if connector.generation == nil {
		connector.generation = newConnectorGeneration()
	}
	if connector.listenerDone != nil {
		connector.mu.Unlock()
		return
	}
	generation := connector.generation
	inCh := connector.inCh
	peerDone := connector.peerDone
	done := make(chan struct{})
	connector.listenerDone = done
	connector.mu.Unlock()

	go func() {
		defer close(done)
		for {
			select {
			case <-generation.stop:
				return
			default:
			}
			select {
			case <-generation.stop:
				return
			case <-peerDone:
				return
			case msg, open := <-inCh:
				if !open {
					return
				}
				if err := nopanic.Run(func() { onMessage(msg.(In)) }); err != nil {
					logger.Error(connector.name, "during onMessage:", err.Error())
				}
			}
		}
	}()
}

func (connector *ServiceConnector[In, Out]) stopListening() {
	connector.opMu.Lock()
	defer connector.opMu.Unlock()
	connector.stopAndJoinCurrent()
}

func (connector *ServiceConnector[In, Out]) stopAndJoinCurrent() {
	connector.mu.RLock()
	generation := connector.generation
	done := connector.listenerDone
	connector.mu.RUnlock()
	generation.close()
	if done != nil {
		<-done
	}
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
	pairs := make([][]string, 0, len(services)/2)
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
		service1, ok := serviceProvider.ByName(pair[0]).(connectable)
		if !ok {
			panic(pair[0] + " can't be connected")
		}
		service2, ok := serviceProvider.ByName(pair[1]).(connectable)
		if !ok {
			panic(pair[1] + " can't be connected")
		}

		srvsStr := strings.Join([]string{pair[0], pair[1]}, ",")
		if service1.inType() != service2.outType() || service1.outType() != service2.inType() {
			panic(srvsStr + " can't be connected")
		}

		done1 := service1.prepareConnection()
		done2 := service2.prepareConnection()
		oneToTwo := make(chan any)
		twoToOne := make(chan any)
		service1.connect(twoToOne, oneToTwo, done2)
		service2.connect(oneToTwo, twoToOne, done1)
		logger.Debug(instance.name, "connected services:", srvsStr)
	}
}

func (instance *mutualConnectableConnector) Name() string {
	return instance.name
}

func (instance *mutualConnectableConnector) Dispose() {}
