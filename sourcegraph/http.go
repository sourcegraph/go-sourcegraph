package sourcegraph

import (
	"net/url"

	"golang.org/x/net/context"
)

// WithHTTPEndpoint returns a copy of parent with the given HTTP API
// endpoint URL set as the HTTP endpoint URL.
//
// Note: all API communication using this package's client is
// performed over gRPC against the GRPCEndpoint; this value is merely
// stored as a convenience for you, not used, by this package.
func WithHTTPEndpoint(parent context.Context, url *url.URL) context.Context {
	return context.WithValue(parent, httpEndpointKey, url)
}

// HTTPEndpoint returns the context's HTTP API endpoint URL that was
// previously set by WithHTTPEndpoint.
//
// Note: all API communication using this package's client is
// performed over gRPC against the GRPCEndpoint; this value is merely
// stored as a convenience for you, not used, by this package.
func HTTPEndpoint(ctx context.Context) *url.URL {
	url, _ := ctx.Value(httpEndpointKey).(*url.URL)
	if url == nil {
		panic("no HTTP API endpoint URL set in context")
	}
	return url
}
