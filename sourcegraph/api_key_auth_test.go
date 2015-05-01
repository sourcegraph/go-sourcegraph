package sourcegraph

import (
	"net/http"
	"testing"
)

type fakeTransport struct {
	req *http.Request
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return nil, nil
}

var secret = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

func TestAPIKeyAuth_HTTP(t *testing.T) {
	const uid = 123

	fakeT := &fakeTransport{}

	authT := &APIKeyAuth{
		Key:       APIKey(secret, uid),
		Transport: fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	authed, authedUID, err := ReadAPIKeyAuth(secret, fakeT.req.Header, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !authed {
		t.Error("!authed")
	}
	if authedUID != uid {
		t.Errorf("got authedUID == %d, want %d", authedUID, uid)
	}
}

func TestAPIKeyAuth_HTTP_fail(t *testing.T) {
	const uid = 123

	fakeT := &fakeTransport{}

	authT := &APIKeyAuth{
		Key:       "123-foo",
		Transport: fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	authed, authedUID, err := ReadAPIKeyAuth(secret, fakeT.req.Header, nil)
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if authed {
		t.Error("authed")
	}
	if authedUID != 0 {
		t.Errorf("got authedUID == %d, want 0", authedUID)
	}
}

func TestAPIKeyAuth_gRPC(t *testing.T) {
	const uid = 123

	authT := &APIKeyAuth{
		Key: APIKey(secret, uid),
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	authed, authedUID, err := ReadAPIKeyAuth(secret, nil, md)
	if err != nil {
		t.Fatal(err)
	}
	if !authed {
		t.Error("!authed")
	}
	if authedUID != uid {
		t.Errorf("got authedUID == %d, want %d", authedUID, uid)
	}
}

func TestAPIKeyAuth_gRPC_fail(t *testing.T) {
	const uid = 123

	authT := &APIKeyAuth{
		Key: "123-foo",
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	authed, authedUID, err := ReadAPIKeyAuth(secret, nil, md)
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if authed {
		t.Error("authed")
	}
	if authedUID != 0 {
		t.Errorf("got authedUID == %d, want 0", authedUID)
	}
}

func TestAPIKey_TextMarshalerUnmarshaler(t *testing.T) {
	ks := APIKey(secret, 123)

	var k1 apiKey
	if err := k1.UnmarshalText([]byte(ks)); err != nil {
		t.Fatal(err)
	}
	if k1.claimedUID != 123 {
		t.Errorf("got claimedUID %d, want 123", k1.claimedUID)
	}

	txt, err := k1.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(txt) != ks {
		t.Errorf("got text %q, want %q", txt, ks)
	}
}
