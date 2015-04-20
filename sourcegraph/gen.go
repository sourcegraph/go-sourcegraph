package sourcegraph

//go:generate protoc -I../../../../github.com/gogo/protobuf/protobuf -I../../../../github.com/gogo/protobuf -I../../../../sourcegraph.com/sourcegraph/go-vcs/vcs -I../../../../sourcegraph.com/sqs/pbtypes -I. --gogo_out=plugins=grpc:. sourcegraph.proto
//go:generate sed -i "s#\\(timestamp\\|void\\).pb#sourcegraph.com/sqs/pbtypes#g" sourcegraph.pb.go
//go:generate sed -i "s#vcs\\.pb#sourcegraph.com/sourcegraph/go-vcs/vcs#g" sourcegraph.pb.go

// NEW mock output style TODO(sqs!nodb-ctx): use this instead of the below when we've merged more stuff
//!go:generate gen-mocks -w -i=.+Serv(er|ice) -o mock -outpkg mock -name_prefix=

//go:generate gen-mocks -w -i=.+Service -outpkg sourcegraph -name_prefix=Mock
