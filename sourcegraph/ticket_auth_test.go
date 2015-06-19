package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
)

func TestTicketAuth_HTTP(t *testing.T) {
	fakeT := &fakeTransport{}

	authT := &TicketAuth{
		SignedTicketStrings: []string{"a", "b"},
		Transport:           fakeT,
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	authT.RoundTrip(req)

	sts, err := ReadTicketAuth(fakeT.req.Header, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sts, authT.SignedTicketStrings) {
		t.Errorf("got ticket strings == %v, want %v", sts, authT.SignedTicketStrings)
	}
}

func TestTicketAuth_gRPC(t *testing.T) {
	authT := &TicketAuth{
		SignedTicketStrings: []string{"a", "b"},
	}

	md, err := authT.GetRequestMetadata(nil)
	if err != nil {
		t.Fatal(err)
	}

	sts, err := ReadTicketAuth(nil, md)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sts, authT.SignedTicketStrings) {
		t.Errorf("got ticket strings == %v, want %v", sts, authT.SignedTicketStrings)
	}
}
