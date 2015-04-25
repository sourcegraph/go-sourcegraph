package sourcegraph

//go:generate protoc -I../../../../ -I../../../../github.com/gogo/protobuf/protobuf -I. --gogo_out=plugins=grpc:. sourcegraph.proto

//go:generate gen-mocks -w -i=.+(Server|Client)$ -o mock -outpkg mock -name_prefix= -no_pass_args=opts

// The pbtypes package selector is emitted as pbtypes1 when more than
// one pbtypes type is used. Fix this up so that goimports works.
//
//go:generate sed -i "s#pbtypes1#pbtypes#g" mock/sourcegraph.pb_mock.go

//go:generate goimports -w mock/sourcegraph.pb_mock.go
