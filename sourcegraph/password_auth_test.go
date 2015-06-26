package sourcegraph

import (
	"net/http"
	"testing"
)

func TestPasswordAuth_HTTP(t *testing.T) {
	fakeT := &fakeTransport{}

	authT := &PasswordAuth{
		Username:  "u",
		Password:  "p",
		Transport: fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	authed, user, pw, err := ReadPasswordAuth(secret, fakeT.req.Header, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !authed {
		t.Error("!authed")
	}
	if want := "u"; user != want {
		t.Errorf("got user == %q, want %q", user, want)
	}
	if want := "p"; pw != want {
		t.Errorf("got pw == %q, want %q", pw, want)
	}
}

func TestPasswordAuth_HTTP_fail(t *testing.T) {
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Header.Set("authorization", "Basic invalid")

	authed, user, pw, err := ReadPasswordAuth(secret, req.Header, nil)
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if authed {
		t.Error("authed")
	}
	if user != "" {
		t.Errorf("got user == %q, want empty", user)
	}
	if pw != "" {
		t.Errorf("got pw == %q, want empty", pw)
	}
}

func TestPasswordAuth_HTTP_none(t *testing.T) {
	req, _ := http.NewRequest("GET", "/foo", nil)

	authed, user, pw, err := ReadPasswordAuth(secret, req.Header, nil)
	if err != nil {
		t.Fatal(err)
	}
	if authed {
		t.Error("authed")
	}
	if user != "" {
		t.Errorf("got user == %q, want empty", user)
	}
	if pw != "" {
		t.Errorf("got pw == %q, want empty", pw)
	}
}

func TestPasswordAuth_gRPC(t *testing.T) {
	authT := &PasswordAuth{
		Username: "u",
		Password: "p",
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	authed, user, pw, err := ReadPasswordAuth(secret, nil, md)
	if err != nil {
		t.Fatal(err)
	}
	if !authed {
		t.Error("!authed")
	}
	if want := "u"; user != want {
		t.Errorf("got user == %q, want %q", user, want)
	}
	if want := "p"; pw != want {
		t.Errorf("got pw == %q, want %q", pw, want)
	}
}

func TestPasswordAuth_gRPC_fail(t *testing.T) {
	authed, user, pw, err := ReadPasswordAuth(secret, nil, map[string]string{passwordAuthMDKey: "invalid"})
	_, isAuthErr := err.(*AuthenticationError)
	if err == nil || !isAuthErr {
		t.Fatalf("got err %v, want AuthenticationError", err)
	}
	if authed {
		t.Error("authed")
	}
	if user != "" {
		t.Errorf("got user == %q, want empty", user)
	}
	if pw != "" {
		t.Errorf("got pw == %q, want empty", pw)
	}
}

func TestPasswordAuth_gRPC_none(t *testing.T) {
	authed, user, pw, err := ReadPasswordAuth(secret, nil, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if authed {
		t.Error("authed")
	}
	if user != "" {
		t.Errorf("got user == %q, want empty", user)
	}
	if pw != "" {
		t.Errorf("got pw == %q, want empty", pw)
	}
}
