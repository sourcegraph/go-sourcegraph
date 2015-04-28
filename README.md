# go-sourcegraph [![Build Status](https://travis-ci.org/sourcegraph/go-sourcegraph.png?branch=master)](https://travis-ci.org/sourcegraph/go-sourcegraph)

[Sourcegraph](https://sourcegraph.com) API client library for [Go](http://golang.org).

**Work in progress. If you want to use this, [post an issue](https://github.com/sourcegraph/go-sourcegraph/issues) or contact us [@srcgraph](https://twitter.com/srcgraph).**

## Development

### Protocol buffers

This repository uses the `sourcegraph/sourcegraph.proto`
[protocol buffers](https://developers.google.com/protocol-buffers/)
definition file to generate Go structs as well as
[gRPC](http://grpc.io) clients and servers for various service
interfaces.

### First-time installation of protobuf tools

You need to install and run the protobuf compiler before you can
regenerate Go code after you change the `sourcegraph.proto` file.

1. **Install protoc**, the protobuf compiler. Find more details at the [protobuf README](https://github.com/google/protobuf)).

   ```
   git clone https://github.com/google/protobuf.git
   cd protobuf
   ./autogen.sh
   ./configure --enable-static && make && sudo make install
   ```
 
   Then make sure the `protoc` binary is in your `$PATH`.
1. **Install [gogo/protobuf](https://github.com/gogo/protobuf)**, on the `proto3` branch.

   ```
   go get -u -a github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}
   cd $GOPATH/github.com/gogo/protobuf
   git checkout proto3
   go install ./...
   ```

### Regenerating Go code after changing `sourcegraph.proto`

1. In `go-sourcegraph` (this repository), run:

   ```
   go generate ./...
   ```

   You can ignore warnings about "No syntax specified for the proto file." These are caused by old protobuf definition files that don't explicitly specify the new proto3 syntax, but they are harmless.
