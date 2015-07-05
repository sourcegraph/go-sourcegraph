package sourcegraph

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
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

// Credentials authenticate gRPC and HTTP requests made by an API
// client.
type Credentials interface {
	oauth2.TokenSource
}

// WithCredentials returns a copy of the parent context that uses cred
// as the credentials for future API clients constructed using this
// context (with NewClientFromContext). It replaces (shadows) any
// previously set credentials in the context.
func WithCredentials(parent context.Context, cred Credentials) context.Context {
	return context.WithValue(parent, credentialsKey, cred)
}

// CredentialsFromContext returns the credentials (if any) previously
// set in the context by WithCredentials.
func CredentialsFromContext(ctx context.Context) Credentials {
	cred, ok := ctx.Value(credentialsKey).(Credentials)
	if !ok {
		return nil
	}
	return cred
}

var maxDialTimeout = 3 * time.Second

// NewClientFromContext returns a Sourcegraph API client configured
// using the context (e.g., authenticated using the context's
// credentials).
var NewClientFromContext = func(ctx context.Context) *Client {
	transport := keepAliveTransport

	opts := []grpc.DialOption{
		grpc.WithCodec(GRPCCodec),
	}

	grpcEndpoint := GRPCEndpoint(ctx)
	if grpcEndpoint.Scheme == "https" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	if cred := CredentialsFromContext(ctx); cred != nil {
		// oauth2.NewClient retrieves the underlying transport from
		// its passed-in context, so we need to create a dummy context
		// using that transport.
		ctxWithTransport := context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: transport})
		transport = oauth2.NewClient(ctxWithTransport, cred).Transport

		// Use contextCredentials instead of directly using the cred
		// so that we can use different credentials for the same
		// connection (in the pool).
		opts = append(opts, grpc.WithPerRPCCredentials(contextCredentials{}))
	}

	// Dial timeout is the lesser of the ctx deadline or
	// maxDialTimeout.
	var timeout time.Duration
	if d, ok := ctx.Deadline(); ok && time.Now().Add(maxDialTimeout).After(d) {
		timeout = d.Sub(time.Now())
	} else {
		timeout = maxDialTimeout
	}
	opts = append(opts, grpc.WithTimeout(timeout))

	conn, err := pooledGRPCDial(grpcEndpoint.Host, opts...)
	if err != nil {
		panic(err)
	}
	c := NewClient(&http.Client{Transport: transport}, conn)
	c.BaseURL = HTTPEndpoint(ctx)
	return c
}

type contextCredentials struct{}

func (contextCredentials) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	if cred := CredentialsFromContext(ctx); cred != nil {
		s := credentials.TokenSource{TokenSource: cred}
		return s.GetRequestMetadata(ctx)
	}
	return nil, nil
}
