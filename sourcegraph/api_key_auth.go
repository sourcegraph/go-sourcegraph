package sourcegraph

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// APIKeyAuth is an HTTP transport and gRPC credential provider that
// authenticates requests with a Sourcegraph API key.
//
// For HTTP/1, it adds HTTP Basic authentication headers to requests.
type APIKeyAuth struct {
	Key string // API key (encodes UID and verifier)

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

func (t APIKeyAuth) NewTransport(underlying http.RoundTripper) http.RoundTripper {
	// Non-pointer method, so we don't modify.
	t.Transport = underlying
	return t
}

// RoundTrip implements the RoundTripper interface.
func (t APIKeyAuth) RoundTrip(req *http.Request) (*http.Response, error) {
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
	req.SetBasicAuth(keyAuthMDKey, t.Key)

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
func (t *APIKeyAuth) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	return map[string]string{keyAuthMDKey: t.Key}, nil
}

// keyAuthMDKey is the gRPC metadata key and HTTP Basic Auth username
// for the API key in authenticated requests.
const keyAuthMDKey = "x-sourcegraph-key"

// ReadAPIKeyAuth reads the client's provided API key from the
// request.
//
// Exactly one of hdr and md must be set. The func takes both
// arguments to avoid the confusion of having one func for reading
// HTTP/1 credentials and another func for reading gRPC credentials.
func ReadAPIKeyAuth(hdr http.Header, md metadata.MD) (key string, err error) {
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
				return "", err
			}
			parts := strings.SplitN(string(val), ":", 2)
			if len(parts) != 2 {
				return "", errors.New("invalid HTTP basic auth header")
			}
			if parts[0] != keyAuthMDKey { // No auth attempted (different scheme).
				return "", nil
			}
			return parts[1], nil
		}
		return "", nil

	case md != nil:
		val, ok := md[keyAuthMDKey]
		if !ok {
			return "", nil
		}
		return val, nil
	}

	return "", nil
}

func VerifyAPIKey(secret []byte, key string) (authed bool, authedUID int, err error) {
	var k apiKey
	if err := k.UnmarshalText([]byte(key)); err != nil {
		return false, 0, &AuthenticationError{Kind: "key", Err: err}
	}

	wantKey := APIKey(secret, k.claimedUID)
	if len(wantKey) == len(key) && subtle.ConstantTimeCompare([]byte(wantKey), []byte(key)) == 1 {
		return true, k.claimedUID, nil
	}
	return false, 0, &AuthenticationError{Kind: "key", Err: errors.New("API key failed verification")}
}

// APIKey constructs an API key.
func APIKey(secret []byte, uid int) string {
	if secret == nil {
		panic("secret must be set")
	}
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(strconv.Itoa(uid))); err != nil {
		panic(err.Error())
	}
	ks, err := (apiKey{claimedUID: uid, verifier: []byte(mac.Sum(nil))}).MarshalText()
	if err != nil {
		panic(err)
	}
	return string(ks)
}

type apiKey struct {
	claimedUID int // the claimed UID must be verified with VerifyAPIKey!
	verifier   []byte
}

func (k apiKey) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d-%s", k.claimedUID, base64.StdEncoding.EncodeToString(k.verifier))), nil
}

func (k *apiKey) UnmarshalText(text []byte) error {
	i := bytes.Index(text, []byte{'-'})
	if i < 1 || i >= len(text)-1 {
		return errors.New("malformatted API key")
	}

	claimedUID, err := strconv.Atoi(string(text[:i]))
	if err != nil {
		return err
	}
	k.claimedUID = claimedUID

	vb := text[i+1:]
	verifier := make([]byte, base64.StdEncoding.DecodedLen(len(vb)))
	vlen, err := base64.StdEncoding.Decode(verifier, vb)
	if err != nil {
		return err
	}
	k.verifier = verifier[:vlen]
	return nil
}
