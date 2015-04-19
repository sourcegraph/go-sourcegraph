package sourcegraph

//!go:generate protoc -I../../../../github.com/gogo/protobuf/protobuf -I../../../../github.com/gogo/protobuf -I. --gogo_out=plugins=grpc:. repos.proto

//go:generate gen-mocks -w -i=.+Serv(er|ice) -o mock -outpkg mock -name_prefix=
