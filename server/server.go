package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/util/addr"
	"github.com/sumlookup/mini/util/meta"
	mnet "github.com/sumlookup/mini/util/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	defaultContentType = "application/grpc"
)

type RegisterImplementation func(s *grpc.Server)

type Server struct {
	Id      string
	Name    string
	Options Options

	// GRPCServer is actual grpc server
	GRPCServer *grpc.Server

	// Holds the service registration
	RegistryService *registry.Service
	// Server Metadata
	Metadata map[string]string
	// Port that has been selected as a server port. This will happen before connection
	Port int
	// Available Handlers
	handlers map[string]Handler
	sync.RWMutex
	wg *sync.WaitGroup
}

func NewServer(opts ...Option) *Server {

	options := newOptions(opts...)

	s := &Server{
		Id:       generateID(options.ServiceName),
		Options:  options,
		Name:     options.ServiceName,
		handlers: make(map[string]Handler),
	}

	s.createGrpcServer()
	return s
}

func (s *Server) Server() *grpc.Server {
	return s.GRPCServer
}

// GetPort returns the grpc server port.
func (s *Server) GetPort() int {
	return s.Port
}

func (s *Server) ServeGRPC(host string, port int) error {
	//err := s.createSubscriptionHandlers()
	//if err != nil {
	//	return err
	//}

	listener, err := s.Options.Transport.Listen(mnet.HostPort(host, port))
	if err != nil {
		return err
	}
	log.Infof("[grpc] Serving on %s", fmt.Sprintf("%s:%v", host, port))
	return s.Server().Serve(listener)
}

// createGrpcServer creates and runs a blocking gRPC server
func (s *Server) createGrpcServer() {

	log.Debugf("Adding %v unary interceptors", len(s.Options.ServerOptions.UnaryInts))
	s.Options.ServerOptions.GRPCOptions = append(s.Options.ServerOptions.GRPCOptions, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(s.Options.ServerOptions.UnaryInts...)))

	log.Debugf("Adding %v stream interceptors", len(s.Options.ServerOptions.StreamInts))
	s.Options.ServerOptions.GRPCOptions = append(s.Options.ServerOptions.GRPCOptions, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(s.Options.ServerOptions.StreamInts...)))

	// set tls if exists
	if s.Options.TLSConfig != nil {
		s.Options.ServerOptions.GRPCOptions = append(s.Options.ServerOptions.GRPCOptions, grpc.Creds(credentials.NewTLS(s.Options.TLSConfig)))
		log.Infof("running with TLS")
	}
	// set the GRPCServer
	s.GRPCServer = grpc.NewServer(
		s.Options.ServerOptions.GRPCOptions...,
	)
}

func (s *Server) AddHandler(h interface{}, opts ...HandlerOption) {
	handler := newGRPCHandler(h, opts...)
	s.handlers[handler.GetName()] = handler
}

// Run registers the service and runs it
func (s *Server) Run() error {

	var serviceRegistry *registry.Service

	var err error
	port := s.Options.ServerOptions.Port

	// find open port it's a port 0. The reason for that is we need to know the port
	// before we start the service so we can register it.
	// this might be refactored in the future. I'm not sure whethere the service
	// registration should happen before we start the service.
	if port == 0 && s.Options.Transport.String() != "memory" {
		port, err = freeport.GetFreePort()
		if err != nil {
			log.Fatal("Could not obtain free port, ", err)
		}
		s.Options.ServerOptions.Port = port
	}

	host := s.Options.ServerOptions.Host
	s.Port = port

	// register the service if the registry is in place
	if s.Options.Registry != nil {

		// extract the ip. This is required so that other services know where this specific service is palced
		ip, err := addr.Extract(host)
		if err != nil {
			return err
		}

		// make copy of metadata
		md := meta.Copy(s.Metadata)

		// Registry service definition. This structure is something that registry takes and
		// uses to register the service.
		address := mnet.HostPort(ip, port)

		log.Debugf("setting up registry, node address %s", address)
		node := &registry.Node{
			Id:       s.Id,
			Address:  address,
			Metadata: md,
		}

		node.Metadata["registry"] = s.Options.Registry.String()
		node.Metadata["protocol"] = "grpc" // we don't have anything else for the moment

		var handlerList []string

		for n, e := range s.handlers {
			// Only advertise non internal handlers
			if !e.GetOptions().Internal {
				handlerList = append(handlerList, n)
			}
		}

		// All the endpoints go to the registry
		endpoints := make([]*registry.Endpoint, 0, len(handlerList))
		for _, n := range handlerList {
			endpoints = append(endpoints, s.handlers[n].GetEndpoints()...)
		}

		// register the service
		serviceRegistry = &registry.Service{
			Name:      s.Options.ServiceName,
			Version:   s.Options.Version,
			Nodes:     []*registry.Node{node},
			Endpoints: endpoints,
		}

		s.RegistryService = serviceRegistry

		log.Infof("%s registry, registering service %s", s.Options.Registry.String(), serviceRegistry.Name)
		err = s.Options.Registry.Register(serviceRegistry)
		if err != nil {
			return err
		}
	} else {
		log.Debugf("no registry set for %s", s.Options.ServiceName)
	}

	go s.signalHandler()
	err = s.ServeGRPC(host, port)
	s.disconnect()
	return err
}

// Stop allows to stop the server gracefully
func (s *Server) Stop() {
	log.Infof("grpc requested server stop")
	s.disconnect()
	log.Debugf("%s grpc initiating graceful stop", s.Options.ServiceName)
	s.GRPCServer.GracefulStop()
}

func (s *Server) signalHandler() {
	sigs := make(chan os.Signal, 1)
	///done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		s.disconnect()
		if os.Getenv("ENV") != "test" {
			os.Exit(0)
		}
	}()
}

// disconnect handles the deregistration of the Registry and Subscriber disconnection
func (s *Server) disconnect() {

	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if s.Options.Registry != nil {
		log.Debugf("grpc deregistering")
		err := s.Options.Registry.Deregister(s.RegistryService,
			registry.DeregisterContext(ctxt),
		)

		if err != nil {
			log.Errorf("[grpc] Could not deregister service. %s from %s: %s", s.Options.ServiceName, s.Options.Registry.String(), err.Error())
		}
	}
}

// newGRPCCodec: checks if codec is defined for given content type
// and returns appropriate codec from defaultGRPCCodecs
func (s *Server) newGRPCCodec(contentType string) (encoding.Codec, error) {
	codecs := make(map[string]encoding.Codec)
	if s.Options.Context != nil {
		if v, ok := s.Options.Context.Value(codecsKey{}).(map[string]encoding.Codec); ok && v != nil {
			codecs = v
		}
	}
	if c, ok := codecs[contentType]; ok {
		return c, nil
	}
	if c, ok := defaultGRPCCodecs[contentType]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("grpc unsupported Content-Type: %s", contentType)
}

// The generated service ID is used for service recognition and mDNS lookup.
// In the RFC https://datatracker.ietf.org/doc/html/rfc6762#page-62 there is
//
//	In the case of DNS label lengths, the stated limit is 63 bytes.  As
//	   with the total name length, this limit is exactly one less than a
//	   power of two.  This label length limit also excludes the label
//	   length byte at the start of every label. Including that extra
//	   byte, a 63-byte label takes 64 bytes of space in memory or in a DNS
//	   message.
func generateID(n string) string {
	uid := uuid.New()
	i := len(n)
	if i > 31 {
		i = 31
	}
	hash := md5.Sum([]byte(n + "-" + uid.String()))
	return string(n[0:i]) + "-" + hex.EncodeToString(hash[:])
}
