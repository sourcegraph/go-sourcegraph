package sourcegraph

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// RegisteredClientAuth is an HTTP transport and gRPC credential
// provider that authenticates requests with a Sourcegraph registered
// API client ID and secret.
//
// For HTTP/1, it adds HTTP authentication headers to requests.
type RegisteredClientAuth struct {
	// Credentials are the registered API client's credentials.
	Credentials RegisteredClientCredentials

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// NewTransport returns a shallow copy of this transport that uses the
// given underlying transport.
func (t RegisteredClientAuth) NewTransport(underlying http.RoundTripper) http.RoundTripper {
	// Non-pointer method, so we don't modify.
	t.Transport = underlying
	return t
}

// RoundTrip implements the RoundTripper interface.
func (t RegisteredClientAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	var transport http.RoundTripper
	if t.Transport != nil {
		transport = t.Transport
	} else {
		transport = http.DefaultTransport
	}

	data, err := t.Credentials.MarshalText()
	if err != nil {
		return nil, err
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = cloneRequest(req)
	req.Header.Add("authorization", registeredClientAuthMDKey+" "+string(data))

	// Make the HTTP request.
	return transport.RoundTrip(req)
}

// GetRequestMetadata implements gRPC's credentials.Credentials
// interface.
func (t *RegisteredClientAuth) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	data, err := t.Credentials.MarshalText()
	if err != nil {
		return nil, err
	}
	return map[string]string{registeredClientAuthMDKey: string(data)}, nil
}

// registeredClientAuthMDKey is the gRPC metadata key and HTTP Basic Auth username
// for the API key in authenticated requests.
const registeredClientAuthMDKey = "x-sourcegraph-registered-client"

// ReadRegisteredClientAuth returns the authenticated registered API
// client for the request. If no authentication is attempted, it
// returns (nil, nil). If authentication fails, a non-nil error is
// returned. If authentication is attempted with valid headers or
// metadata, a non-nil RegisteredClientCredentials is returned, but
// THIS FUNCTION DOES NOT CHECK WHETHER IT IS VALID.
//
// Exactly one of hdr and md must be set. The func takes both
// arguments to avoid the confusion of having one func for reading
// HTTP/1 credentials and another func for reading gRPC credentials.
func ReadRegisteredClientAuth(hdr http.Header, md metadata.MD) (*RegisteredClientCredentials, error) {
	decode := func(data string) (*RegisteredClientCredentials, error) {
		var cred RegisteredClientCredentials
		if err := cred.UnmarshalText([]byte(data)); err != nil {
			return nil, &AuthenticationError{Kind: "RegisteredClientAuth", Err: err}
		}
		return &cred, nil
	}

	switch {
	case (hdr != nil && md != nil) || (hdr == nil && md == nil):
		panic("exactly one of hdr and md must be set")

	case hdr != nil:
		for _, rawval := range hdr[http.CanonicalHeaderKey("authorization")] {
			if !strings.HasPrefix(rawval, registeredClientAuthMDKey+" ") {
				continue
			}
			rawval = strings.TrimSpace(strings.TrimPrefix(rawval, registeredClientAuthMDKey+" "))
			return decode(rawval)
		}
		return nil, nil

	case md != nil:
		txt, ok := md[registeredClientAuthMDKey]
		if !ok {
			return nil, nil
		}
		return decode(txt)
	}

	panic("unreachable")
}
