package service

import (
	"fmt"
	//"honnef.co/go/tools/config"

	//"github.com/sumlookup/mini/loglevel"
	//"github.com/sumlookup/mini/registry"
	//"github.com/sumlookup/mini/registry/mdns"
	//"github.com/sumlookup/mini/registry/memory"
	//"github.com/opentracing/opentracing-go"
	//"io"
	//"regexp"
	"time"

	//"github.com/sumlookup/mini/dependency"
	//"github.com/sumlookup/mini/util/env"

	log "github.com/sirupsen/logrus"
	"github.com/sumlookup/mini/builder"
	client "github.com/sumlookup/mini/client"
	"github.com/sumlookup/mini/util/env"
	//"github.com/sumlookup/mini/config"
	"github.com/sumlookup/mini/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	Name   string
	Srv    *server.Server
	Closer *Closer
	//Provider ServiceProvider
	start bool

	transport string
	registry  string
	//selector  string
}

//const ServiceProviderPluginName = "CreateServiceProvider"

// ServiceProviderPluginEntry is a definition of the function that returns the service provider interface
//type ServiceProviderPluginEntry func(provider dependency.DependencyProvider) Service

//func init() {
//	loglevel.SetLogging()
//}

//type ServiceProvider interface {
//	InitService() func(app *Service) error
//	GetServiceClient(cl *client.Client) interface{}
//	GetServiceProtoName() string
//	GetServiceName() string
//	//GetCli() ([]*cli.Command, error)
//	//GetDependencyProvider() dependency.DependencyProvider
//
//	// GetProjectName returns the project name. ex: rosi, sirius. List of available project names in tplib.c9helpers
//	// project name is required for conststency reasons when building topic names and many others
//	GetProjectName() string
//}

type Closer struct {
	closers []func()
}

var Version string
var Commit string

// NewService creates a new service with the default environment added
func NewService(name, transport, registry string) *Service {
	if Version == "" {
		Version = "dev"
	}

	// map of function that is executed when closing the service
	closer := &Closer{
		closers: []func(){},
	}

	//var cfg config.Config
	var srv *Service
	start := false
	var grpcSrv *server.Server

	srvOpts, err := BuildServerOptions(name, transport, registry)
	if err != nil {
		log.Error(err)
	}

	start = true

	// override with the ones passed as argument
	//s := newServiceOptions(sopt, sopts...)

	// create the server
	grpcSrv = server.NewServer(srvOpts...)
	// register reflection

	serv := grpcSrv.Server()
	reflection.Register(serv)

	srv = &Service{
		Name: name,
		//ServiceOptions: sopt,
		Srv:    grpcSrv,
		Closer: closer,
		//Config:         cfg,
		start:     start,
		transport: transport,
		registry:  registry,
	}

	return srv
}

func (s *Service) Version() string {
	return Version
}

func (s *Service) Commit() string {
	return Commit
}

func (s *Service) AddHandler(h interface{}, opts ...server.HandlerOption) {
	if h != nil && s.Srv != nil {
		s.Srv.AddHandler(h, opts...)
	}
}

func (s *Service) Run() error {

	if s.start {

		e := env.New()
		log.Infof("service %s, env: %s, version %s", s.Name, e.GetEnv(), Version)
		return s.Srv.Run()
	}
	return nil
}

func (s *Service) GetSrv() *server.Server {
	return s.Srv
}

func (s *Service) Close() {
	s.Srv.Stop()
}

//func (s *Service) SetServiceProvider(p ServiceProvider) {
//	s.Provider = p
//}

//	func (s *Service) NewClient(targetService string, opts ...client.Option) *client.Client {
//		// merge options passed in service. take opts argument as priority
//		//opts = append(s.ServiceOptions.ClientOptions, opts...)
//		return client.NewClient(s.Name, opts...)
//	}
//
// todo: move transort / reg to be set on service and re-used
func (s *Service) Client(selector string, opts ...client.Option) *client.Client {
	//opts = append(s.Srv.Options.s, opts...)
	return client.NewClient(s.Name, s.transport, s.registry, selector, opts...)
}

func (s *Service) Server() *grpc.Server {
	return s.Srv.Server()
}

//func (s *Service) GetConfig() config.Config {
//	return s.Config
//}

func (c *Closer) Append(closer func()) {
	c.closers = append(c.closers, closer)
}

// GetPort is blocking until the server port is set
// Background: server.Run is blocking therefore accessing generated port
// will be 0 unless we wait for the server to start and then ask for port
// by default this function will wait 120 seconds for the server to be available
// after that it will return an error so that the code which uses this function can fallback
// on other methods
func (s *Service) GetPort() (int, error) {

	// return if it's already set
	if s.GetSrv().GetPort() != 0 {
		return s.GetSrv().GetPort(), nil
	}

	// this needs to wait for the port allocation because grpc is blocking
	ticker := time.NewTicker(200 * time.Millisecond)
	done := make(chan bool)

	timeout := 120
	// port -1 and wait for the port to be set, return when done
	port := -1
	i := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				i++
				if s.GetSrv().GetPort() != 0 {
					port = s.GetSrv().GetPort()
					done <- true
					return
				}

				if i >= 5*timeout { //120 seconds
					done <- true
					log.Warn("grpc server port is not available")
					return
				}
			}
		}
	}()

	<-done
	ticker.Stop()
	if port == -1 {
		log.Warn("can't wait no longer for the port, falling back to the standard service resolution")
		return 0, fmt.Errorf("can't retrieve port in the last %v seconds", timeout)
	}
	return port, nil
}

func BuildServerOptions(serviceName, tr, reg string) ([]server.Option, error) {

	var so []server.Option

	log.Infof("building server options for %s", serviceName)
	// assign config

	transport := builder.BuildTransport(tr)
	log.Infof("server %s transport: %s", serviceName, transport.String())
	err := transport.Init()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	registry := builder.BuildRegistry(reg)
	log.Infof("server %s registry: %s", serviceName, registry.String())

	so = append(so,
		server.WithRegistry(registry),
		//server.WithPort(c.Int("port")),
		server.WithTransport(transport),
		server.ServiceName(serviceName),
	)

	// TODO IMPLEMENT INTERCEPTORS -- TRACER / LOGGERS ETC..
	var ssi []grpc.StreamServerInterceptor
	var usi []grpc.UnaryServerInterceptor

	so = append(so,
		server.UnaryInterceptor(usi...),
		server.StreamInterceptor(ssi...),
	)

	log.Debugf("built %v server options", len(so))
	return so, nil
}
