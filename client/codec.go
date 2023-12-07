package client

import (
	"github.com/sumlookup/mini/codec"
	raw "github.com/sumlookup/mini/codec/bytes"
	"github.com/sumlookup/mini/codec/grpc"
	"github.com/sumlookup/mini/codec/json"
	"github.com/sumlookup/mini/codec/proto"
)

type Codecs map[string]codec.NewCodec

var (
	DefaultContentType = "application/protobuf"
	DefaultCodecs      = Codecs{
		"application/grpc":         grpc.NewCodec,
		"application/protobuf":     proto.NewCodec,
		"application/json":         json.NewCodec,
		"application/octet-stream": raw.NewCodec,
	}
)
