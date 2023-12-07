// Package static provides a static resolver which returns the name/ip passed in without any change
package fixed

import (
	"fmt"
	"os"

	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/selector"
)

const (
	ENV_FIXED_SELECTOR_DOMAIN_NAME = "FIXED_SELECTOR_DOMAIN_NAME"
	ENV_FIXED_SELECTOR_PORT_NUMBER = "FIXED_SELECTOR_PORT_NUMBER"
	DEFAULT_PORT_NUMBER            = "8080"
)

type staticSelector struct {
	address       string
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
	node := &registry.Node{
		Id:      service,
		Address: fmt.Sprintf("%v", s.address),
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
	return "fixed"
}

func NewSelector(opts ...selector.Option) selector.Selector {

	// Build a new selector
	s := &staticSelector{
		address:       "",
		envDomainName: os.Getenv(ENV_FIXED_SELECTOR_DOMAIN_NAME),
		envPortNumber: os.Getenv(ENV_FIXED_SELECTOR_PORT_NUMBER),
	}

	// Add the dns domain-name (if one was specified by an env-var):
	if s.envDomainName != "" {
		s.address += fmt.Sprintf("%v", s.envDomainName)
	}

	// Either add the default port-number, or override with one specified by an env-var:
	if s.envPortNumber == "" {
		s.address += fmt.Sprintf(":%v", DEFAULT_PORT_NUMBER)
	} else {
		s.address += fmt.Sprintf(":%v", s.envPortNumber)
	}

	return s
}
