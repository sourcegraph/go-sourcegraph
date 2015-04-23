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

func TestKeyAuth_HTTP(t *testing.T) {
	const uid = 123

	fakeT := &fakeTransport{}

	authT := &KeyAuth{
		UID:       uid,
		Key:       AuthKey(secret, uid),
		Transport: fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	authed, authedUID, err := ReadKeyAuth(secret, fakeT.req.Header, nil)
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

func TestKeyAuth_HTTP_fail(t *testing.T) {
	const uid = 123

	fakeT := &fakeTransport{}

	authT := &KeyAuth{
		UID:       uid,
		Key:       "foo",
		Transport: fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	authed, authedUID, err := ReadKeyAuth(secret, fakeT.req.Header, nil)
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

func TestKeyAuth_gRPC(t *testing.T) {
	const uid = 123

	authT := &KeyAuth{
		UID: uid,
		Key: AuthKey(secret, uid),
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	authed, authedUID, err := ReadKeyAuth(secret, nil, md)
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

func TestKeyAuth_gRPC_fail(t *testing.T) {
	const uid = 123

	authT := &KeyAuth{
		UID: uid,
		Key: "foo",
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	authed, authedUID, err := ReadKeyAuth(secret, nil, md)
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
