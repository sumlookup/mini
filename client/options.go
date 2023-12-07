package client

import (
	"context"
	"crypto/tls"
	"github.com/sumlookup/mini/selector"
	"github.com/sumlookup/mini/transport"
	"google.golang.org/grpc"
	"time"
)

type Options struct {
	DialOptions           DialOptions
	UnaryInts             []grpc.UnaryClientInterceptor
	StreamInts            []grpc.StreamClientInterceptor
	Selector              selector.Selector
	SelectOptions         []selector.SelectOption
	TLSConfig             *tls.Config
	Codecs                Codecs
	Context               context.Context
	ConnectionMaxAttempts int
	ConnectionTicker      time.Duration
	ConnectionAttempts    bool
	Transport             transport.Transport
	GrpcConnection        grpc.ClientConnInterface
	ContentType           string
	HostOverride          string
}

type DialOption grpc.DialOption
type DialOptions []grpc.DialOption

// Option used by the Client
type Option func(*Options)

func NewOptions(options ...Option) Options {

	opts := Options{
		DialOptions: DialOptions{
			//grpc.WithTimeout(DefaultDialTimeout),
			//grpc.WithBlock(),
		},
		ContentType:           DefaultContentType,
		Codecs:                DefaultCodecs,
		ConnectionMaxAttempts: 3,
		ConnectionTicker:      2,
		//HealthCheckTicker:     5,
		//ConnectionHealthCheck: false, // this is potentially harmfull as it will keep to call the service even if it has been closed
		ConnectionAttempts: true,
	}

	for _, o := range options {
		o(&opts)
	}

	return opts
}

func WithGrpcConnection(c grpc.ClientConnInterface) Option {
	return func(o *Options) {
		o.GrpcConnection = c
	}
}

func WithDialOptions(n DialOptions) Option {
	return func(o *Options) {
		o.DialOptions = append(o.DialOptions, n...)
	}
}

//func WithConnectionHealthCheck(h bool) Option {
//	return func(o *Options) {
//		o.ConnectionHealthCheck = h
//	}
//}

func WithConnectionAttempts(h bool) Option {
	return func(o *Options) {
		o.ConnectionAttempts = h
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

func ContentType(ct string) Option {
	return func(o *Options) {
		o.ContentType = ct
	}
}

// Select is used to select a node to route a request to
func Selector(s selector.Selector) Option {
	return func(o *Options) {
		o.Selector = s
	}
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// Specify broker
//func Broker(b broker.Broker) Option {
//	return func(o *Options) {
//		o.Broker = b
//	}
//}

func WithTransport(tr transport.Transport) Option {
	return func(o *Options) {
		o.Transport = tr
	}
}

// UnaryInterceptor
func UnaryInterceptor(r ...grpc.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.UnaryInts = append(o.UnaryInts, r...)
	}
}

func StreamInterceptor(r ...grpc.StreamClientInterceptor) Option {
	return func(o *Options) {
		o.StreamInts = append(o.StreamInts, r...)
	}
}

func WithMaxConnectionAttempts(i int) Option {
	return func(o *Options) {
		o.ConnectionMaxAttempts = i
	}
}

func WithHostOverride(host string) Option {
	return func(o *Options) {
		o.HostOverride = host
	}
}
