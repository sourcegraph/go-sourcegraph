package sourcegraph

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// PasswordAuth is an HTTP transport and gRPC credential provider that
// authenticates requests with a username and password.
//
// For HTTP/1, it adds HTTP Basic authentication headers to requests.
type PasswordAuth struct {
	Username string
	Password string

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

func (t PasswordAuth) NewTransport(underlying http.RoundTripper) http.RoundTripper {
	// Non-pointer method, so we don't modify.
	t.Transport = underlying
	return t
}

// RoundTrip implements the RoundTripper interface.
func (t PasswordAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	var transport http.RoundTripper
	if t.Transport != nil {
		transport = t.Transport
	} else {
		transport = http.DefaultTransport
	}

	// To set extra querystring params, we must make a copy of the
	// Request so that we don't modify the Request we were given. This
	// is required by the specification of http.RoundTripper.
	req = cloneRequest(req)
	req.SetBasicAuth(t.Username, t.Password)

	// Make the HTTP request.
	return transport.RoundTrip(req)
}

// GetRequestMetadata implements gRPC's credentials.Credentials
// interface.
func (t *PasswordAuth) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		passwordAuthMDKey: base64.StdEncoding.EncodeToString([]byte(t.Username + ":" + t.Password)),
	}, nil
}

// passwordAuthMDKey is the gRPC metadata key for the encoded username
// and password in PasswordAuth-authenticated requests.
const passwordAuthMDKey = "x-sourcegraph-username-password"

// ReadPasswordAuth reads the username/password auth data from the
// HTTP Basic auth header or the x-sourcegraph-username-password gRPC
// metadata key.
//
// Exactly one of hdr and md must be set. The func takes both
// arguments to avoid the confusion of having one func for reading
// HTTP/1 credentials and another func for reading gRPC credentials.
func ReadPasswordAuth(secret []byte, hdr http.Header, md metadata.MD) (authed bool, login, password string, err error) {
	var b64Encoded string

	switch {
	case hdr != nil && md != nil:
		panic("exactly one of hdr and md must be set")

	case hdr != nil:
		for _, rawval := range hdr[http.CanonicalHeaderKey("authorization")] {
			if !strings.HasPrefix(rawval, "Basic ") {
				continue
			}
			b64Encoded = strings.TrimSpace(strings.TrimPrefix(rawval, "Basic "))
			break
		}

	case md != nil:
		var ok bool
		b64Encoded, ok = md[passwordAuthMDKey]
		if !ok {
			return false, "", "", nil
		}
	}

	if b64Encoded == "" {
		return false, "", "", nil
	}

	// Parse HTTP basic auth header.
	val, err := base64.StdEncoding.DecodeString(b64Encoded)
	if err != nil {
		return false, "", "", &AuthenticationError{Kind: "password", Err: err}
	}
	parts := strings.SplitN(string(val), ":", 2)
	if len(parts) != 2 {
		return false, "", "", &AuthenticationError{
			Kind: "password",
			Err:  errors.New("invalid encoded password auth credentials"),
		}
	}
	return true, parts[0], parts[1], nil
}
