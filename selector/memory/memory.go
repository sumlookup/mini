// Package static provides a static resolver which returns the name/ip passed in without any change
package static

import (
	"fmt"

	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/selector"
)

type memorySelector struct {
}

func (s *memorySelector) Init(opts ...selector.Option) error {
	return nil
}

func (s *memorySelector) Options() selector.Options {
	return selector.Options{}
}

func (s *memorySelector) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	node := &registry.Node{
		Id:      service,
		Address: fmt.Sprintf("%v:%v", service, 0),
	}

	return func() (*registry.Node, error) {
		return node, nil
	}, nil
}

func (s *memorySelector) Mark(service string, node *registry.Node, err error) {
	return
}

func (s *memorySelector) Reset(service string) {
	return
}

func (s *memorySelector) Close() error {
	return nil
}

func (s *memorySelector) String() string {
	return "memory"
}

func NewSelector(opts ...selector.Option) selector.Selector {
	return &memorySelector{}
}
