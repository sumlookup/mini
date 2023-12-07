package builder

import (
	log "github.com/sirupsen/logrus"
	"github.com/sumlookup/mini/registry"
	"github.com/sumlookup/mini/registry/mdns"
	"github.com/sumlookup/mini/registry/memory"
	"github.com/sumlookup/mini/selector"
	sm "github.com/sumlookup/mini/selector/memory"
	sr "github.com/sumlookup/mini/selector/registry"
	st "github.com/sumlookup/mini/selector/static"
	"github.com/sumlookup/mini/transport"
	transportGrpc "github.com/sumlookup/mini/transport/grpc"
	transportMemory "github.com/sumlookup/mini/transport/memory"
)

func BuildTransport(transportName string) transport.Transport {

	var tr transport.Transport

	switch transportName {
	case "grpc":
		tr = transportGrpc.NewTransport()
	case "memory":
		tr = transportMemory.NewTransport()
	default:
		tr = transportGrpc.NewTransport()
	}

	return tr
}

func BuildSelector(s string, r registry.Registry) selector.Selector {
	var sel selector.Selector
	switch s {
	case "registry":
		sel = sr.NewSelector(selector.Registry(r))
	case "static":
		sel = st.NewSelector()
	case "memory":
		sel = sm.NewSelector()
	}

	return sel
}

func BuildRegistry(r string) registry.Registry {
	var reg registry.Registry
	switch r {
	case "mdns":
		reg = mdns.NewRegistry()
	case "memory":
		reg = memory.NewRegistry()
	default:
		log.Warnf("Defaulted to registry : mdns")
		reg = mdns.NewRegistry()
	}
	return reg
}
