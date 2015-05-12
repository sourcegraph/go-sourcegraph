package sourcegraph

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"strings"
	"sync"

	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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

func TestPerRPCCredentials(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(l); err != nil {
			t.Fatal(err)
		}
	}()
	defer s.TestingCloseConns()

	var ms testMetaServer
	RegisterMetaServer(s, &ms)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			time.Sleep(time.Duration(i%10) * 5 * time.Millisecond)

			key := fmt.Sprintf("key%d", i)

			ctx := context.Background()
			ctx = WithGRPCEndpoint(ctx, &url.URL{Host: l.Addr().String()})
			ctx = WithHTTPEndpoint(ctx, &url.URL{Scheme: "http", Host: l.Addr().String()})
			ctx = WithClientCredentials(ctx, &APIKeyAuth{Key: key})
			ctx = metadata.NewContext(ctx, metadata.MD{"want-x-sourcegraph-key": key})
			c := NewClientFromContext(ctx)
			defer c.Close()
			if _, err := c.Meta.Status(ctx, &pbtypes.Void{}); err != nil {
				t.Fatal(err)
			}
		}(i)
	}
	wg.Wait()

	out, err := exec.Command("netstat", "-ntap").CombinedOutput()
	if err == nil {
		lines := bytes.Split(out, []byte("\n"))
		var conns, timeWaits int
		addr := strings.Replace(l.Addr().String(), "[::]", "::1", 1)
		for _, line := range lines {
			if bytes.Contains(line, []byte(addr)) {
				conns++
				if bytes.Contains(line, []byte("TIME_WAIT")) {
					timeWaits++
				}
			}
		}
		t.Logf("lingering connections count: %d", conns)
		t.Logf("         in TIME_WAIT state: %d", timeWaits)
		t.Log("(ideally, there should be 0 lingering connections)")
	} else {
		t.Logf("warning: error running `netstat -ntap` to check # of TIME_WAIT conns: %s", err)
	}
}

type testMetaServer struct {
	MetaServer
}

func (s *testMetaServer) Status(ctx context.Context, _ *pbtypes.Void) (*ServerStatus, error) {
	md, _ := metadata.FromContext(ctx)
	if want, got := md["want-x-sourcegraph-key"], md["x-sourcegraph-key"]; got != want {
		return nil, grpc.Errorf(codes.Unknown, "got x-sourcegraph-key %q, want %q", got, want)
	}
	return &ServerStatus{}, nil
}
