package sourcegraph

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type contextKey int

const (
	grpcEndpointKey contextKey = iota
	httpEndpointKey
	credentialsKey
)

// WithGRPCEndpoint returns a copy of parent whose clients (obtained
// using FromContext) communicate with the given gRPC API endpoint
// URL.
func WithGRPCEndpoint(parent context.Context, url *url.URL) context.Context {
	return context.WithValue(parent, grpcEndpointKey, url)
}

// GRPCEndpoint returns the context's gRPC endpoint URL that was
// previously configured using WithGRPCEndpoint.
func GRPCEndpoint(ctx context.Context) *url.URL {
	url, _ := ctx.Value(grpcEndpointKey).(*url.URL)
	if url == nil {
		panic("no gRPC API endpoint URL set in context")
	}
	return url
}

// WithHTTPEndpoint returns a copy of parent whose clients (obtained
// using FromContext) communicate with the given HTTP API endpoint
// URL.
func WithHTTPEndpoint(parent context.Context, url *url.URL) context.Context {
	return context.WithValue(parent, httpEndpointKey, url)
}

// HTTPEndpoint returns the context's HTTP API endpoint URL that was
// previously configured using WithHTTPEndpoint.
func HTTPEndpoint(ctx context.Context) *url.URL {
	url, _ := ctx.Value(httpEndpointKey).(*url.URL)
	if url == nil {
		panic("no HTTP API endpoint URL set in context")
	}
	return url
}

// Credentials is implemented by authentication providers that provide
// both gRPC and HTTP auth (e.g., API keys, tickets).
type Credentials interface {
	// The TokenSource adds authentication info to gRPC calls.
	credentials.Credentials

	// NewTransport creates a new HTTP transport that adds
	// authentication info to outgoing HTTP requests and calls the
	// underlying transport. It MUST NOT modify the Credentials object
	// (e.g., it should return a copy of it with its Transport field
	// set to the underlying transport).
	NewTransport(underlying http.RoundTripper) http.RoundTripper
}

// WithClientCredentials adds cred as a credential provider for future
// API clients constructed using this context (with FromContext).
func WithClientCredentials(parent context.Context, cred Credentials) context.Context {
	creds := clientCredentialsFromContext(parent)
	return context.WithValue(parent, credentialsKey, append([]Credentials{cred}, creds...))
}

func clientCredentialsFromContext(ctx context.Context) []Credentials {
	creds, ok := ctx.Value(credentialsKey).([]Credentials)
	if !ok {
		return nil
	}
	return creds
}

// NewClientFromContext returns a Sourcegraph API client configured
// using the context (e.g., authenticated using the context's
// credentials (actor & tickets)).
var NewClientFromContext = func(ctx context.Context) *Client {
	transport := http.DefaultTransport

	// TODO(sqs!): make appdash work with gRPC
	// if spanID := traceutil.SpanID(r); spanID.Trace != 0 {
	// 	// Add appdash tracing info.
	// 	transport = &httptrace.Transport{
	// 		Recorder:  traceutil.Recorder(spanID),
	// 		Transport: transport,
	// 	}
	// }

	opts := []grpc.DialOption{
		grpc.WithCodec(GRPCCodec),
	}

	grpcEndpoint := GRPCEndpoint(ctx)
	if grpcEndpoint.Scheme == "https" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	for _, cred := range clientCredentialsFromContext(ctx) {
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
		transport = cred.NewTransport(transport)
	}

	conn, err := grpc.Dial(grpcEndpoint.Host, opts...)
	if err != nil {
		panic(err)
	}
	c := NewClient(&http.Client{Transport: transport}, conn)
	c.BaseURL = HTTPEndpoint(ctx)
	return c
}
