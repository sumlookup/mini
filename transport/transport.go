// Package transport is an interface for synchronous connection based communication
package transport

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

// Transport is an interface which is used for communication between
// services. It uses connection based socket send/recv semantics and
// has various implementations; http, grpc, quic.
type Transport interface {
	Init(...Option) error
	Options() Options
	Dial(addr string, opts ...grpc.DialOption) (grpc.ClientConnInterface, error)
	Listen(addr string, opts ...ListenOption) (net.Listener, error)
	String() string
}

type Message struct {
	Header map[string]string
	Body   []byte
}

type Socket interface {
	Close() error
	Local() string
	Remote() string
}

type Client interface {
	Socket
}

type Listener interface {
	Addr() string
	Close() error
	Accept() error
}

type Option func(*Options)

type DialOption func(*DialOptions)

type ListenOption func(*ListenOptions)

var (
	DefaultDialTimeout = time.Second * 5
)
