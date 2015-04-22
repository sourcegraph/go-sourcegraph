package sourcegraph

import "github.com/gogo/protobuf/proto"

// GogoCodec uses gogo/protobuf instead of golang/protobuf to encode
// gRPC messages. It's needed because we use gogo-specific options
// (nullable, embed, etc.)
var GogoCodec gogoCodec

type gogoCodec struct{}

func (gogoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (gogoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (gogoCodec) String() string { return "gogoprotobuf" }
