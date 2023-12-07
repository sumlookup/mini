package client

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sumlookup/mini/builder"
	"github.com/sumlookup/mini/codec"
	"github.com/sumlookup/mini/selector"
	"google.golang.org/grpc"
	"time"
)

var DefaultClientOptions []Option

type Client struct {
	Id             string
	Options        Options
	GRPCConnection grpc.ClientConnInterface
	ServiceName    string
	//once           atomic.Value
	seq uint64
}

// NewClient creates a new client with the default options
func NewClient(clientName, transport, registry, selector string, opts ...Option) *Client {
	var co []Option
	var err error

	co, err = BuildClientOptions(transport, registry, selector)
	if err != nil {
		log.Error(err)
	}

	co = append(opts, co...)
	log.Infof("%s creatig new client", clientName)
	return New(co...)
}

func New(opts ...Option) *Client {

	// object options
	o := NewOptions(opts...)

	// Client
	c := &Client{
		Id:      fmt.Sprintf("client-%s", uuid.New().String()),
		Options: o,
		seq:     0,
	}

	// pass the grpc connection to the client if we have one in options
	if o.GrpcConnection != nil {
		c.GRPCConnection = o.GrpcConnection
	}

	//c.once.Store(false)

	return c
}

func (c *Client) Connect(serviceName string) grpc.ClientConnInterface {
	c.ServiceName = serviceName
	log.Infof("dialing %s", c.ServiceName)
	// create connection and resuse in the future
	if c.GRPCConnection == nil {
		// connect will attempt to connect, if the server is not found it iwll
		// make multiple attempts to connect
		err := c.connectToService()
		if err != nil {
			log.Error(err)
			return nil
		}

	} else {
		log.Warnf("grpc client not creating connection to %s. It already exists", c.ServiceName)
	}

	if c.GRPCConnection == nil {
		log.Fatalf("Tried to retrieve a connection that is not there")
	}

	return c.GRPCConnection
}

func BuildClientOptions(tr, reg, sel string) ([]Option, error) {

	var co []Option

	if len(DefaultClientOptions) > 0 {
		return DefaultClientOptions, nil
	}

	log.Infof("building options for client")
	registry := builder.BuildRegistry(reg)
	log.Infof("client registry: %s", registry.String())
	transport := builder.BuildTransport(tr)
	log.Infof("client transport: %s", tr)
	err := transport.Init()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	co = append(co,
		Selector(builder.BuildSelector(sel, registry)),
		WithTransport(transport),
	)

	// TODO: add tracers and logger interceptors...
	//co = append(co,
	//	UnaryInterceptor(),
	//	StreamInterceptor(),
	//)
	log.Infof("built %v client options", len(co))
	DefaultClientOptions = co
	return co, nil
}

// attempts to connect to the service, if Options.ConnectionAttempts is true then it will
// also try to connect for the period of time if the service is not available on the other side
// this is useful for client load balancing in local development
func (c *Client) connectToService() error {

	conn, err := c.createConnection()

	if c.Options.ConnectionAttempts {
		// if error, try more
		if err != nil {
			log.Infof("attempting to connect to %s for the next %v seconds", c.ServiceName, int(c.Options.ConnectionTicker)*c.Options.ConnectionMaxAttempts)
			ticker := time.NewTicker(c.Options.ConnectionTicker * time.Second)
		L:
			for i := 0; i < c.Options.ConnectionMaxAttempts; i++ {
				select {
				case <-ticker.C:

					conn, err = c.createConnection()
					if err != nil {
						log.Debugf(err.Error())
					} else {
						break L
					}
				}
			}

			ticker.Stop()
		}

		// that's it.. we give up
		if conn == nil {
			return fmt.Errorf("could not create client for %s. Exhausted connection attempts: %v", c.ServiceName, c.Options.ConnectionMaxAttempts)
		}

	} else {
		log.Debug("did not attempt to reconnect")
		if err != nil {
			return fmt.Errorf("could not create client for %s: %s", c.ServiceName, err.Error())
		}
	}

	c.GRPCConnection = conn
	return nil
}

// createConnection to the service
func (c *Client) createConnection() (grpc.ClientConnInterface, error) {

	// return existing connection if availble
	if c.GRPCConnection != nil {
		log.Debugf("grpc client returning existig connection to %s", c.ServiceName)
		return c.GRPCConnection, nil
	}

	host, err := c.getServiceHost()
	if err != nil {
		log.Warnf("can't get the host, %s", err.Error())
		return nil, err
	}

	log.Infof("client dials %s at %s using %s, %v unary interceptors", c.ServiceName, host, c.Options.Transport.String(), len(c.Options.UnaryInts))

	options := append(c.Options.DialOptions, []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(c.Options.UnaryInts...),
		grpc.WithChainStreamInterceptor(c.Options.StreamInts...),
	}...)

	conn, err := c.Options.Transport.Dial(host, options...)
	if err != nil {
		log.Warnf("could not establish connection to service %s: %s", host, err.Error())
		return nil, err
	}

	if conn == nil {
		return nil, fmt.Errorf("Could not connect to the service %s at %s", c.ServiceName, host)
	}

	return conn, nil
}

// getServiceHost: Getting service name from registered node
func (c *Client) getServiceHost() (string, error) {

	// force host override for situations where we have to connect to something else
	if c.Options.HostOverride != "" {
		log.Infof("overriding host name to %s", c.Options.HostOverride)
		return c.Options.HostOverride, nil
	}

	host := c.ServiceName
	if c.Options.Selector != nil {
		log.Debugf("grpc client using selector: %s", c.Options.Selector.String())
		next, err := c.next(c.ServiceName)
		if err != nil {
			return "", fmt.Errorf("grpc client, %s selector could not select the connection to %s: %s", c.Options.Selector.String(), c.ServiceName, err.Error())
		}

		// retrieve the node details
		node, err := next()
		if err != nil {
			return "", fmt.Errorf("grpc client selector could not retrieve node address to %s: %s", c.ServiceName, err.Error())
		}

		host = node.Address
	}

	log.Debugf("grpc client host used for connection to %s: %s", c.ServiceName, host)
	return host, nil
}

func (c *Client) next(serviceName string) (selector.Next, error) {

	if c.Options.Selector == nil {
		log.Warnf("Selecector not defined")
		return nil, nil
	}

	next, err := c.Options.Selector.Select(serviceName, c.Options.SelectOptions...)

	if err != nil {
		if err == selector.ErrNotFound {
			return nil, err
		}
		return nil, err
	}
	return next, nil
}

// newCodec: checks if codec is defined for given content type
// and returns appropriate codec from defaultCodecs
func (c *Client) newCodec(contentType string) (codec.NewCodec, error) {
	if c, ok := c.Options.Codecs[contentType]; ok {
		return c, nil
	}
	if cf, ok := DefaultCodecs[contentType]; ok {
		return cf, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type:  " + contentType)
}
