// Package static provides a static resolver which returns the name/ip passed in without any change
package static

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/selector"
	"github.com/sumlookup/mini/util/env"
)

const (
	ENV_STATIC_SELECTOR_DOMAIN_NAME = "STATIC_SELECTOR_DOMAIN_NAME"
	ENV_STATIC_SELECTOR_SUFFIX      = "STATIC_SELECTOR_SUFFIX"
	ENV_STATIC_SELECTOR_ENVMOD      = "STATIC_SELECTOR_ENVMOD"
	ENV_STATIC_SELECTOR_PORT_NUMBER = "STATIC_SELECTOR_PORT_NUMBER"
	DEFAULT_PORT_NUMBER             = "8080"
)

type staticSelector struct {
	addressSuffix string
	envDomainName string
	envPortNumber string
}

func (s *staticSelector) Init(opts ...selector.Option) error {
	return nil
}

func (s *staticSelector) Options() selector.Options {
	return selector.Options{}
}

func (s *staticSelector) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	service, layer := s.processSuffix(service)

	address := service
	if !s.isLocalhost(service) {
		address = fmt.Sprintf("%v%v%v%v", service, layer, s.envDomainName, s.envPortNumber)
	} else {
		// localhost
		address = fmt.Sprintf("%v%v", service, s.envPortNumber)
	}

	node := &registry.Node{
		Id:      service,
		Address: address,
	}

	return func() (*registry.Node, error) {
		return node, nil
	}, nil
}

func (s *staticSelector) Mark(service string, node *registry.Node, err error) {
	return
}

func (s *staticSelector) Reset(service string) {
	return
}

func (s *staticSelector) Close() error {
	return nil
}

func (s *staticSelector) String() string {
	return "static"
}

func (s *staticSelector) isLocalhost(service string) bool {
	return strings.Contains(service, "127") || strings.Contains(service, "localhost")
}

// simple function for custom layer
func (s *staticSelector) processSuffix(service string) (string, string) {
	fmt.Printf("processSuffix @@ - svc - %v \n", service)
	layer := s.addressSuffix

	// allow the service name to be used as a namespace identifier
	// core-role = role.core.svc.cluster.local

	if strings.Contains(layer, "splitservice") {
		split := strings.Split(service, "-")
		if len(split) >= 2 {
			service = strings.Join(split[1:], "-")
			layer = fmt.Sprintf(".%v", split[0])
		}

		// allows the service to be addressed in the parameterised namespace, ex:
		// connection.core-uat.svc.cluster.local
		if os.Getenv(ENV_STATIC_SELECTOR_ENVMOD) == "true" {
			e := env.New()
			layer = fmt.Sprintf("%s-%s", layer, e.GetEnv())
		}

		// allows the service to ignore the namespaces and use service name only
		// ex: core-role = role.default.svc.cluser.local, role.local, etc
	} else if strings.Contains(layer, "name") {
		split := strings.Split(service, "-")
		if len(split) >= 2 {
			service = strings.Join(split[1:], "-")
			layer = ""
		}
	} else if strings.Contains(layer, "env") {
		service = service + "." + os.Getenv("ENV")
		layer = ""
	} else if strings.Contains(layer, "direct") {
		return service, ""
	} else if layer == "" {
		return service, ""
	} else {
		log.Fatalf("static selector misconfigured, layer = '%s', service '%s' ", layer, service)
	}

	return service, layer
}

func NewSelector(opts ...selector.Option) selector.Selector {

	// Build a new
	s := &staticSelector{
		addressSuffix: os.Getenv(ENV_STATIC_SELECTOR_SUFFIX),
		envDomainName: os.Getenv(ENV_STATIC_SELECTOR_DOMAIN_NAME),
		envPortNumber: os.Getenv(ENV_STATIC_SELECTOR_PORT_NUMBER),
	}

	// Add the dns domain-name (if one was specified by an env-var):
	if s.envDomainName != "" {
		s.envDomainName = fmt.Sprintf(".%v", s.envDomainName)
	}

	// Either add the default port-number, or override with one specified by an env-var:
	if s.envPortNumber == "" {
		s.envPortNumber = fmt.Sprintf(":%v", DEFAULT_PORT_NUMBER)
	} else {
		s.envPortNumber = fmt.Sprintf(":%v", s.envPortNumber)
	}

	return s
}
