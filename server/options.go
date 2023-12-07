package server

import (
	"context"
	"crypto/tls"
	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/transport"
	"google.golang.org/grpc"
)

type Option func(*Options)
type codecsKey struct{}

type ServerOptions struct {
	GRPCOptions []grpc.ServerOption
	UnaryInts   []grpc.UnaryServerInterceptor
	StreamInts  []grpc.StreamServerInterceptor
	Host        string
	Port        int
}

type Options struct {
	ServiceName   string
	Version       string
	OwnerName     string
	OwnerEmail    string
	ServerOptions *ServerOptions
	Registry      registry.Registry
	TLSConfig     *tls.Config
	Context       context.Context
	Transport     transport.Transport
	Tracer        string
}

type Handler interface {
	GetName() string
	GetHandler() interface{}
	GetEndpoints() []*registry.Endpoint
	GetOptions() HandlerOptions
}

type Message interface {
	// Topic of the messagee
	Topic() string
	// The decoded payload value
	Payload() interface{}
	// The content type of the payload
	ContentType() string
	// The raw headers of the message
	Header() map[string]string
	// The raw body of the message
	Body() []byte
	// Codec used to decode the message
	//Codec() codec.Reader
}

type HandlerOption func(*HandlerOptions)

type HandlerOptions struct {
	Internal bool
	Metadata map[string]map[string]string
}

func EndpointMetadata(name string, md map[string]string) HandlerOption {
	return func(o *HandlerOptions) {
		o.Metadata[name] = md
	}
}

func InternalHandler(b bool) HandlerOption {
	return func(o *HandlerOptions) {
		o.Internal = b
	}
}

// newOptions creates default Options and adds the passed Option(s)
func newOptions(opt ...Option) Options {

	// default options
	opts := Options{
		Version: "v0.0.1",
		ServerOptions: &ServerOptions{
			Port:        0,
			Host:        "0.0.0.0",
			GRPCOptions: []grpc.ServerOption{},
			UnaryInts:   []grpc.UnaryServerInterceptor{},
			StreamInts:  []grpc.StreamServerInterceptor{},
		},
	}

	// load the options from the parameter
	for _, o := range opt {
		o(&opts)
	}

	return opts
}

// UnaryInterceptor
func UnaryInterceptor(r ...grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.ServerOptions.UnaryInts = append(o.ServerOptions.UnaryInts, r...)
	}
}

func StreamInterceptor(r ...grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.ServerOptions.StreamInts = append(o.ServerOptions.StreamInts, r...)
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.ServerOptions.Port = port
	}
}

func WithHost(host string) Option {
	return func(o *Options) {
		o.ServerOptions.Host = host
	}
}

func WithRegistry(reg registry.Registry) Option {
	return func(o *Options) {
		o.Registry = reg
	}
}

func Version(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func ServiceName(name string) Option {
	return func(o *Options) {
		o.ServiceName = name
	}
}

func WithTransport(tr transport.Transport) Option {
	return func(o *Options) {
		o.Transport = tr
	}
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// Tracer name
func Tracer(name string) Option {
	return func(o *Options) {
		o.Tracer = name
	}
}
