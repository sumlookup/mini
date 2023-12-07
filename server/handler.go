package server

import (
	"fmt"
	"github.com/sumlookup/mini/registry"
	"reflect"
	"strings"
)

type rpcHandler struct {
	Name      string `json:"name"`
	Handler   interface{}
	Endpoints []*registry.Endpoint
	Opts      HandlerOptions
}

func newGRPCHandler(handler interface{}, opts ...HandlerOption) Handler {
	options := HandlerOptions{
		Metadata: make(map[string]map[string]string),
	}

	for _, o := range opts {
		o(&options)
	}

	typ := reflect.TypeOf(handler)
	hdlr := reflect.ValueOf(handler)
	name := reflect.Indirect(hdlr).Type().Name()

	var endpoints []*registry.Endpoint

	for m := 0; m < typ.NumMethod(); m++ {
		if e := extractEndpoint(typ.Method(m)); e != nil {
			e.Name = name + "." + e.Name
			for k, v := range options.Metadata[e.Name] {
				e.Metadata[k] = v
			}

			endpoints = append(endpoints, e)
		}
	}

	return &rpcHandler{
		Name:      name,
		Handler:   handler,
		Endpoints: endpoints,
		Opts:      options,
	}

}

func (r *rpcHandler) GetName() string {
	return r.Name
}

func (r *rpcHandler) GetHandler() interface{} {
	return r.Handler
}

func (r *rpcHandler) GetEndpoints() []*registry.Endpoint {
	return r.Endpoints
}

func (r *rpcHandler) GetOptions() HandlerOptions {
	return r.Opts
}

func extractValue(v reflect.Type, d int) *registry.Value {
	if d == 3 {
		return nil
	}
	if v == nil {
		return nil
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	arg := &registry.Value{
		Name: v.Name(),
		Type: v.Name(),
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			val := extractValue(f.Type, d+1)
			if val == nil {
				continue
			}

			// if we can find a json tag use it
			if tags := f.Tag.Get("json"); len(tags) > 0 {
				parts := strings.Split(tags, ",")
				if parts[0] == "-" || parts[0] == "omitempty" {
					continue
				}
				val.Name = parts[0]
			} else {
				continue
			}

			arg.Values = append(arg.Values, val)
		}
	case reflect.Slice:
		p := v.Elem()
		if p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		arg.Type = "[]" + p.Name()
	}

	return arg
}

func extractEndpoint(method reflect.Method) *registry.Endpoint {
	if method.PkgPath != "" {
		return nil
	}

	var rspType, reqType reflect.Type
	var streamIn, streamOut bool
	mt := method.Type

	switch mt.NumIn() {
	case 2: // Usually a stream
		reqType = mt.In(1)  // first request in the stream mode is a server
		rspType = mt.Out(0) // this will always be error
		streamIn = true

	case 3:
		reqType = mt.In(2)  // always request (0 = name, 1 = context)
		rspType = mt.Out(0) // always response

		// This is the type of rpc that accepts the request and returns stream.
		// It will have 3 values, func name, request and server response
		if rspType.Kind().String() == "interface" {
			reqType = mt.In(1)
			rspType = mt.In(2)
			streamOut = true
		}

	default:
		return nil
	}

	request := extractValue(reqType, 0)

	//Always process the first response
	response := extractValue(rspType, 0)

	ep := &registry.Endpoint{
		Name:     method.Name,
		Request:  request,
		Response: response,
		Metadata: make(map[string]string),
	}

	// set endpoint metadata for stream
	if streamIn {
		ep.Metadata = map[string]string{
			"streamIn": fmt.Sprintf("%v", streamIn),
			"stream":   "true",
		}
	}

	if streamOut {
		ep.Metadata = map[string]string{
			"streamOut": fmt.Sprintf("%v", streamOut),
			"stream":    "true",
		}
	}

	return ep
}

//
//func extractSubValue(typ reflect.Type) *registry.Value {
//	var reqType reflect.Type
//	switch typ.NumIn() {
//	case 1:
//		reqType = typ.In(0)
//	case 2:
//		reqType = typ.In(1)
//	case 3:
//		reqType = typ.In(2)
//	default:
//		return nil
//	}
//	return extractValue(reqType, 0)
//}
