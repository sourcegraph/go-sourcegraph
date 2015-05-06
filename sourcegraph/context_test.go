package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type dummyCredentials struct {
	X int
	credentials.Credentials
}

func (dummyCredentials) NewTransport(underlying http.RoundTripper) http.RoundTripper {
	return http.RoundTripper(nil)
}

func TestWithClientCredentials(t *testing.T) {
	ctx := context.Background()

	if creds := clientCredentialsFromContext(ctx); len(creds) != 0 {
		t.Errorf("got %+v, want empty", creds)
	}

	dummy := dummyCredentials{X: 1}
	wantCreds := []Credentials{dummy}
	ctx = WithClientCredentials(ctx, dummy)
	if creds := clientCredentialsFromContext(ctx); !reflect.DeepEqual(creds, wantCreds) {
		t.Errorf("got %+v, want %+v", creds, wantCreds)
	}

	dummy2 := dummyCredentials{X: 2}
	wantCreds = []Credentials{dummy2, dummy}
	ctx = WithClientCredentials(ctx, dummy2)
	if creds := clientCredentialsFromContext(ctx); !reflect.DeepEqual(creds, wantCreds) {
		t.Errorf("got %+v, want %+v", creds, wantCreds)
	}
}
