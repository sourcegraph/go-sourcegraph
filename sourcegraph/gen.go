package sourcegraph

//go:generate protoc -I../../../../github.com/gogo/protobuf/protobuf -I../../../../github.com/gogo/protobuf -I. --gogo_out=plugins=grpc:. sourcegraph.proto timestamp.proto void.proto

// NEW mock output style TODO(sqs!nodb-ctx): use this instead of the below when we've merged more stuff
//!go:generate gen-mocks -w -i=.+Serv(er|ice) -o mock -outpkg mock -name_prefix=

//go:generate gen-mocks -w -i=.+Service -outpkg sourcegraph -name_prefix=Mock
