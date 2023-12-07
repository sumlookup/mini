// Package grpc provides a grpc transport
package cgrpc

import (
	"context"
	"crypto/tls"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/sumlookup/mini/transport"
	maddr "github.com/sumlookup/mini/util/addr"
	mnet "github.com/sumlookup/mini/util/net"
	mls "github.com/sumlookup/mini/util/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type grpcTransport struct {
	opts transport.Options
}

type grpcTransportListener struct {
	listener net.Listener
	secure   bool
	tls      *tls.Config
}

func getTLSConfig(addr string) (*tls.Config, error) {
	hosts := []string{addr}

	// check if its a valid host:port
	if host, _, err := net.SplitHostPort(addr); err == nil {
		if len(host) == 0 {
			hosts = maddr.IPs()
		} else {
			hosts = []string{host}
		}
	}

	// generate a certificate
	cert, err := mls.Certificate(hosts...)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func (t *grpcTransport) Dial(addr string, dialOptions ...grpc.DialOption) (grpc.ClientConnInterface, error) {

	if t.opts.Secure || t.opts.TLSConfig != nil {
		config := t.opts.TLSConfig
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

	// dial the server
	ctx, cancel := context.WithTimeout(context.Background(), t.opts.Timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, dialOptions...)
	if err != nil {
		log.Errorf("can't dial to %s", addr)
		return nil, err
	}

	return conn, nil
}

func (t *grpcTransport) Listen(addr string, opts ...transport.ListenOption) (net.Listener, error) {
	var options transport.ListenOptions
	for _, o := range opts {
		o(&options)
	}

	ln, err := mnet.Listen(addr, func(addr string) (net.Listener, error) {
		log.Infof("listening on %s", addr)
		return net.Listen("tcp", addr)
	})
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func (t *grpcTransport) Init(opts ...transport.Option) error {
	for _, o := range opts {
		o(&t.opts)
	}
	return nil
}

func (t *grpcTransport) Options() transport.Options {
	return t.opts
}

func (t *grpcTransport) String() string {
	return "grpc"
}

func NewTransport(opts ...transport.Option) transport.Transport {
	log.Info("initialising tcp transport manager")
	var options transport.Options
	options.Timeout = transport.DEFAULT_TIMEOUT
	for _, o := range opts {
		o(&options)
	}
	return &grpcTransport{opts: options}
}
