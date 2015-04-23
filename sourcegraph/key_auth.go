package sourcegraph

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// KeyAuth is an HTTP transport and gRPC credential provider that
// authenticates requests with a Sourcegraph API key.
//
// For HTTP/1, it adds HTTP Basic authentication headers to requests.
type KeyAuth struct {
	UID int    // user ID
	Key string // API key

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *KeyAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	var transport http.RoundTripper
	if t.Transport != nil {
		transport = t.Transport
	} else {
		transport = http.DefaultTransport
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = cloneRequest(req)
	req.SetBasicAuth(strconv.Itoa(t.UID), t.Key)

	// Make the HTTP request.
	return transport.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request. The clone is a
// shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}

// GetRequestMetadata implements gRPC's credentials.Credentials
// interface.
func (t *KeyAuth) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	return map[string]string{keyAuthMDKey: strconv.Itoa(t.UID) + ":" + t.Key}, nil
}

// keyAuthMDKey is the gRPC metadata key used to store the API key in
// authenticated requests.
const keyAuthMDKey = "key-auth"

// ReadKeyAuth returns the authenticated UID for the request. If no
// authentication is attempted, it returns (false, 0, nil). If
// authentication fails, a non-nil error is returned. If
// authentication succeeds, authed=true.
//
// Exactly one of hdr and md must be set. The func takes both
// arguments to avoid the confusion of having one func for reading
// HTTP/1 credentials and another func for reading gRPC credentials.
func ReadKeyAuth(secret []byte, hdr http.Header, md metadata.MD) (authed bool, uid int, err error) {
	switch {
	case hdr != nil && md != nil:
		panic("exactly one of hdr and md must be set")

	case hdr != nil:
		for _, rawval := range hdr[http.CanonicalHeaderKey("authorization")] {
			if !strings.HasPrefix(rawval, "Basic ") {
				continue
			}
			rawval = strings.TrimSpace(strings.TrimPrefix(rawval, "Basic "))

			// Parse HTTP basic auth header.
			val, err := base64.StdEncoding.DecodeString(rawval)
			if err != nil {
				return false, 0, err
			}
			parts := strings.SplitN(string(val), ":", 2)
			if len(parts) != 2 {
				return false, 0, errors.New("invalid HTTP basic auth header")
			}
			uidStr, key := parts[0], parts[1]
			uid, err = strconv.Atoi(uidStr)
			if err != nil {
				return false, 0, err
			}
			return verifyAuthKey(secret, uid, key)
		}
		return false, 0, nil

	case md != nil:
		val, ok := md[keyAuthMDKey]
		if !ok {
			return false, 0, nil
		}
		parts := strings.SplitN(string(val), ":", 2)
		if len(parts) != 2 {
			return false, 0, errors.New("invalid gRPC key auth metadata entry")
		}
		uidStr, key := parts[0], parts[1]
		uid, err = strconv.Atoi(uidStr)
		if err != nil {
			return false, 0, err
		}
		return verifyAuthKey(secret, uid, key)
	}

	return false, 0, nil
}

func verifyAuthKey(secret []byte, uid int, key string) (authed bool, authedUID int, err error) {
	wantKey := AuthKey(secret, uid)
	if len(wantKey) == len(key) && subtle.ConstantTimeCompare([]byte(wantKey), []byte(key)) == 1 {
		return true, uid, nil
	}
	return false, 0, &AuthenticationError{Kind: "key", Err: errors.New("API key failed verification")}
}

// AuthKey constructs an auth key.
func AuthKey(secret []byte, uid int) string {
	if secret == nil {
		panic("secret must be set")
	}
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(strconv.Itoa(uid))); err != nil {
		panic(err.Error())
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
