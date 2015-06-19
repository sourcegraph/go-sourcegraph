package sourcegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "sourcegraph-client/" + libraryVersion
)

// A Client communicates with the Sourcegraph API. All communication
// is done using gRPC over HTTP/2 except for BuildData (which uses
// HTTP/1).
type Client struct {
	// Services used to communicate with different parts of the Sourcegraph API.
	Accounts            AccountsClient
	Builds              BuildsClient
	BuildData           BuildDataService
	Defs                DefsClient
	Deltas              DeltasClient
	Markdown            MarkdownClient
	Meta                MetaClient
	MirrorRepos         MirrorReposClient
	MirroredRepoSSHKeys MirroredRepoSSHKeysClient
	Orgs                OrgsClient
	People              PeopleClient
	RepoBadges          RepoBadgesClient
	RepoStatuses        RepoStatusesClient
	RepoTree            RepoTreeClient
	Repos               ReposClient
	Search              SearchClient
	Units               UnitsClient
	UserAuth            UserAuthClient
	Users               UsersClient

	// Base URL for HTTP/1.1 requests, which should have a trailing slash.
	BaseURL *url.URL

	// User agent used for HTTP/1.1 requests to the Sourcegraph API.
	UserAgent string

	// HTTP client used to communicate with the Sourcegraph API.
	httpClient *http.Client

	// gRPC client connection used to communicate with the Sourcegraph
	// API.
	Conn *grpc.ClientConn
}

// NewClient returns a Sourcegraph API client. The gRPC conn is used
// for all services except for BuildData (which uses the
// httpClient). If httpClient is nil, http.DefaultClient is used.
func NewClient(httpClient *http.Client, conn *grpc.ClientConn) *Client {
	c := new(Client)

	// HTTP/1
	if httpClient == nil {
		cloned := *http.DefaultClient
		cloned.Transport = keepAliveTransport
		httpClient = &cloned
	}
	c.httpClient = httpClient
	c.BaseURL = &url.URL{Scheme: "https", Host: "sourcegraph.com", Path: "/api/"}
	c.UserAgent = userAgent
	c.BuildData = &buildDataService{c}

	// gRPC (HTTP/2)
	c.Conn = conn
	c.Accounts = NewAccountsClient(conn)
	c.Builds = NewBuildsClient(conn)
	c.Defs = NewDefsClient(conn)
	c.Deltas = NewDeltasClient(conn)
	c.Markdown = NewMarkdownClient(conn)
	c.Meta = NewMetaClient(conn)
	c.MirrorRepos = NewMirrorReposClient(conn)
	c.MirroredRepoSSHKeys = NewMirroredRepoSSHKeysClient(conn)
	c.Orgs = NewOrgsClient(conn)
	c.People = NewPeopleClient(conn)
	c.RepoBadges = NewRepoBadgesClient(conn)
	c.RepoStatuses = NewRepoStatusesClient(conn)
	c.RepoTree = NewRepoTreeClient(conn)
	c.Repos = NewReposClient(conn)
	c.Search = NewSearchClient(conn)
	c.Units = NewUnitsClient(conn)
	c.UserAuth = NewUserAuthClient(conn)
	c.Users = NewUsersClient(conn)

	return c
}

// Router is used to generate URLs for the Sourcegraph API.
var Router = router.NewAPIRouter(nil)

// ResetRouter clears and reconstructs the preinitialized API
// router. It should be called after setting an router.ExtraConfig
// func but only during init time.
func ResetRouter() {
	Router = router.NewAPIRouter(nil)
}

// URL generates a URL for the given route, route variables, and
// querystring options. Unless you explicitly set a Host, Scheme,
// and/or Port on Router, the returned URL will contain only path and
// querystring components (and will not be an absolute URL).
func URL(route string, routeVars map[string]string, opt interface{}) (*url.URL, error) {
	rt := Router.Get(route)
	if rt == nil {
		return nil, fmt.Errorf("no Sourcegraph API route named %q", route)
	}

	routeVarsList := make([]string, 2*len(routeVars))
	i := 0
	for name, val := range routeVars {
		routeVarsList[i*2] = name
		routeVarsList[i*2+1] = val
		i++
	}
	url, err := rt.URL(routeVarsList...)
	if err != nil {
		return nil, err
	}

	if opt != nil {
		err = addOptions(url, opt)
		if err != nil {
			return nil, err
		}
	}

	return url, nil
}

// URL generates the absolute URL to the named Sourcegraph API endpoint, using the
// specified route variables and query options.
func (c *Client) URL(route string, routeVars map[string]string, opt interface{}) (*url.URL, error) {
	url, err := URL(route, routeVars, opt)
	if err != nil {
		return nil, err
	}

	// make the route URL path relative to BaseURL by trimming the leading "/"
	url.Path = strings.TrimPrefix(url.Path, "/")

	// make the route URL path relative to BaseURL's path and not the path parent
	baseURL := *c.BaseURL
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path = baseURL.Path + "/"
	}

	// make the URL absolute
	url = baseURL.ResolveReference(url)

	return url, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client. Relative
// URLs should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// newResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) Response {
	if r == nil {
		return nil
	}
	return &HTTPResponse{Response: r}
}

// HTTPResponse is a wrapped HTTP response from the Sourcegraph API with
// additional Sourcegraph-specific response information parsed out. It
// implements Response.
type HTTPResponse struct {
	*http.Response
}

// TotalCount implements Response.
func (r *HTTPResponse) TotalCount() int {
	tc := r.Header.Get("x-total-count")
	if tc == "" {
		return -1
	}
	n, err := strconv.Atoi(tc)
	if err != nil {
		return -1
	}
	return n
}

// Response is a response from the Sourcegraph API. When using the HTTP API,
// API methods return *HTTPResponse values that implement Response.
type Response interface {
	// TotalCount is the total number of items in the resource or result set
	// that exist remotely. Only a portion of the total may be in the response
	// body. If the endpoint did not return a total count, then TotalCount
	// returns -1.
	TotalCount() int
}

// SimpleResponse implements Response.
type SimpleResponse struct {
	Total int // see (Response).TotalCount()
}

func (r *SimpleResponse) TotalCount() int { return r.Total }

type doKey int // sentinel value type for (*Client).Do v parameter

const preserveBody doKey = iota // when passed as v to (*Client).Do, the resp body is neither parsed nor closed

// Do sends an API request and returns the API response.  The API
// response is decoded and stored in the value pointed to by v, or
// returned as an error if an API error has occurred. If v is
// preserveBody, then the HTTP response body is not closed by Do; the
// caller is responsible for closing it.
func (c *Client) Do(req *http.Request, v interface{}) (Response, error) {
	var resp Response
	rawResp, err := c.httpClient.Do(req)
	if rawResp != nil {
		if v != preserveBody && rawResp.Body != nil {
			defer rawResp.Body.Close()
		}
		resp = newResponse(rawResp)
		if err == nil {
			// Don't clobber error from Do, if any (it could be, e.g.,
			// a sentinel error returned by the HTTP client's
			// CheckRedirect func).
			if err := CheckResponse(rawResp); err != nil {
				// even though there was an error, we still return the response
				// in case the caller wants to inspect it further
				return resp, err
			}
		}
	}
	if err != nil {
		return resp, err
	}

	if v != nil {
		if bp, ok := v.(*[]byte); ok {
			*bp, err = ioutil.ReadAll(rawResp.Body)
		} else if v != preserveBody {
			err = json.NewDecoder(rawResp.Body).Decode(v)
		}
	}
	if err != nil {
		return resp, fmt.Errorf("error reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
	}
	return resp, nil
}

// addOptions adds the parameters in opt as URL query parameters to u. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(u *url.URL, opt interface{}) error {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	qs, err := query.Values(opt)
	if err != nil {
		return err
	}

	u.RawQuery = qs.Encode()
	return nil
}

// keepAliveTransport is an http.RoundTripper that uses a larger
// keep-alive pool than the default.
var keepAliveTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,

	// Allow more keep-alive connections per host to avoid
	// ephemeral port exhaustion due to getting stuck in
	// TIME_WAIT. Some systems have a very limited ephemeral port
	// supply (~1024). 20 connections is perfectly reasonable,
	// since this client will only ever hit one host.
	MaxIdleConnsPerHost: 20,
}
