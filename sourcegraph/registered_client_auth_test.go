package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestRegisteredClientAuth_HTTP(t *testing.T) {
	fakeT := &fakeTransport{}

	authT := &RegisteredClientAuth{
		Credentials: RegisteredClientCredentials{ID: "a", Secret: "b"},
		Transport:   fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	cred, err := ReadRegisteredClientAuth(fakeT.req.Header, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cred == nil {
		t.Error("cred == nil")
	}
	if !reflect.DeepEqual(*cred, authT.Credentials) {
		t.Errorf("got cred == %+v, want %+v", *cred, authT.Credentials)
	}
}

func TestRegisteredClientAuth_HTTP_none(t *testing.T) {
	fakeT := &fakeTransport{}

	req, _ := http.NewRequest("GET", "/foo", nil)
	fakeT.RoundTrip(req)

	cred, err := ReadRegisteredClientAuth(fakeT.req.Header, nil)
	if err != nil {
		t.Fatalf("got err %v, want nil", err)
	}
	if cred != nil {
		t.Errorf("got cred == %+v, want nil", *cred)
	}
}

func TestRegisteredClientAuth_HTTP_fail(t *testing.T) {
	fakeT := &fakeTransport{}

	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Header.Set("authorization", "x-sourcegraph-registered-client badformat")
	fakeT.RoundTrip(req)

	cred, err := ReadRegisteredClientAuth(fakeT.req.Header, nil)
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if cred != nil {
		t.Errorf("got cred == %+v, want nil", *cred)
	}
}

func TestRegisteredClientAuth_gRPC(t *testing.T) {
	authT := &RegisteredClientAuth{
		Credentials: RegisteredClientCredentials{ID: "a", Secret: "b"},
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	cred, err := ReadRegisteredClientAuth(nil, md)
	if err != nil {
		t.Fatal(err)
	}
	if cred == nil {
		t.Error("cred == nil")
	}
	if !reflect.DeepEqual(*cred, authT.Credentials) {
		t.Errorf("got cred == %+v, want %+v", *cred, authT.Credentials)
	}
}

func TestRegisteredClientAuth_gRPC_none(t *testing.T) {
	cred, err := ReadRegisteredClientAuth(nil, metadata.MD{})
	if err != nil {
		t.Fatalf("got err %v, want nil", err)
	}
	if cred != nil {
		t.Errorf("got cred == %+v, want nil", *cred)
	}
}

func TestRegisteredClientAuth_gRPC_fail(t *testing.T) {
	cred, err := ReadRegisteredClientAuth(nil, metadata.MD{"x-sourcegraph-registered-client": "badformat"})
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if cred != nil {
		t.Errorf("got cred == %+v, want nil", *cred)
	}
}
