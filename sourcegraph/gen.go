package sourcegraph

//go:generate protoc -I../../../../github.com/gogo/protobuf/protobuf -I../../../../github.com/gogo/protobuf -I. --gogo_out=plugins=grpc:. sourcegraph.proto timestamp.proto void.proto

//go:generate gen-mocks -w -i=.+Serv(er|ice) -o mock -outpkg mock -name_prefix=
