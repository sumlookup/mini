package memory

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/credentials"

	"github.com/sumlookup/mini/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	log "github.com/sirupsen/logrus"
)

var DefaultMemoryTransportManager *MemoryTransportManager

func init() {
	log.Debug("initialising memory transport manager")
	DefaultMemoryTransportManager = NewMemoryTransportManager()
}

type MemoryTransportManager struct {
	lock      sync.RWMutex
	Listeners map[string]*bufconn.Listener
}

func NewMemoryTransportManager() *MemoryTransportManager {
	return &MemoryTransportManager{Listeners: make(map[string]*bufconn.Listener), lock: sync.RWMutex{}}
}

func (m *MemoryTransportManager) AddListener(name string, t *bufconn.Listener) {
	m.lock.Lock()
	defer m.lock.Unlock()
	log.Debugf("adding %s memory listener", name)
	m.Listeners[name] = t
}

func (m *MemoryTransportManager) GetListener(name string) *bufconn.Listener {
	m.lock.Lock()
	defer m.lock.Unlock()
	log.Debugf("getting %s memory listener", name)
	if _, ok := m.Listeners[name]; ok {
		return m.Listeners[name]
	}

	avl := []string{}
	for k := range m.Listeners {
		avl = append(avl, k)
	}
	log.Debugf("%s memory listener not available. %v listeners in pool, %s", name, len(m.Listeners), avl)
	return nil
}

type memoryTransport struct {
	connections   chan net.Conn
	state         chan int
	isStateClosed uint32
	client        net.Conn
	opts          transport.Options
	sync.RWMutex
}

func NewTransport(opts ...transport.Option) transport.Transport {
	var options transport.Options
	for _, o := range opts {
		o(&options)
	}

	return &memoryTransport{
		opts:        options,
		connections: make(chan net.Conn),
		state:       make(chan int),
	}
}

func (t *memoryTransport) Init(opts ...transport.Option) error {
	for _, o := range opts {
		o(&t.opts)
	}
	return nil
}

func (t *memoryTransport) Options() transport.Options {
	return t.opts
}

func (t *memoryTransport) String() string {
	return "memory"
}

func (t *memoryTransport) Listen(addr string, opts ...transport.ListenOption) (net.Listener, error) {
	//ml := NewMemoryListener()
	buffer := 1024
	listener := bufconn.Listen(buffer)
	DefaultMemoryTransportManager.AddListener(addr, listener)
	return listener, nil
}

func (ml *memoryTransport) Close() error {
	return nil
}

func (ml *memoryTransport) Dial(addr string, dialOptions ...grpc.DialOption) (grpc.ClientConnInterface, error) {

	listener := DefaultMemoryTransportManager.GetListener(addr)
	if listener == nil {
		return nil, fmt.Errorf("Memory transport manager can't find %s listener", addr)
	}

	dialOptions = append(dialOptions, []grpc.DialOption{grpc.WithContextDialer(func(a context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure()}...)

	if ml.opts.Secure || ml.opts.TLSConfig != nil {
		config := ml.opts.TLSConfig
		if config == nil {
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		tlsCredentials := credentials.NewTLS(config)
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(tlsCredentials))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	ctx, cancel := context.WithTimeout(context.Background(), transport.DefaultDialTimeout)
	defer cancel()

	// connect
	conn, err := grpc.DialContext(ctx, "", dialOptions...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type MemoryListener struct {
	connections   chan net.Conn
	state         chan int
	isStateClosed uint32
}

func NewMemoryListener() *MemoryListener {
	ml := &MemoryListener{}
	ml.connections = make(chan net.Conn)
	ml.state = make(chan int)
	return ml
}

func (ml *MemoryListener) Accept() (net.Conn, error) {
	log.Debug("Waiting for connections")
	select {
	case newConnection := <-ml.connections:
		return newConnection, nil
	case <-ml.state:
		return nil, errors.New("Listener closed")
	}
}

func (ml *MemoryListener) Close() error {
	log.Debugf("Closing memory listener")
	if atomic.CompareAndSwapUint32(&ml.isStateClosed, 0, 1) {
		close(ml.state)
	}
	return nil
}

func (ml *MemoryListener) Dial(network, addr string) (net.Conn, error) {

	select {
	case <-ml.state:
		return nil, errors.New("Listener closed")
	default:
	}

	//Create an in memory transport
	serverSide, clientSide := net.Pipe()
	//Pass half to the server
	ml.connections <- serverSide
	//Return the other half to the client
	return clientSide, nil
}

func (ml *MemoryListener) Addr() net.Addr {
	return &Addr{}
}

// net.Addr implementation
type Addr struct{}

func (a *Addr) Network() string {
	return "memory"
}

func (a *Addr) String() string {
	return "memory"
}
