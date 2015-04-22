package sourcegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-querystring/query"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "sourcegraph-client/" + libraryVersion
)

// A Client communicates with the Sourcegraph API.
type Client struct {
	// Services used to communicate with different parts of the Sourcegraph API.
	Builds       BuildsClient
	Deltas       DeltasClient
	Orgs         OrgsClient
	People       PeopleClient
	Repos        ReposClient
	RepoStatuses RepoStatusesClient
	RepoBadges   RepoBadgesClient
	RepoTree     RepoTreeClient
	Search       SearchClient
	Units        UnitsClient
	Users        UsersClient
	Defs         DefsClient
	Markdown     MarkdownClient
	VCS          VCSOpener

	// Base URL for API requests, which should have a trailing slash.
	BaseURL *url.URL

	// User agent used for HTTP requests to the Sourcegraph API.
	UserAgent string

	// HTTP client used to communicate with the Sourcegraph API.
	httpClient *http.Client
}

func NewGRPCClient(conn *grpc.ClientConn) *Client {
	if conn == nil {
		panic("conn == nil")
	}

	c := new(Client)
	c.Builds = NewBuildsClient(conn)
	c.Deltas = NewDeltasClient(conn)
	c.Orgs = NewOrgsClient(conn)
	c.People = NewPeopleClient(conn)
	c.Repos = NewReposClient(conn)
	c.RepoStatuses = NewRepoStatusesClient(conn)
	c.RepoBadges = NewRepoBadgesClient(conn)
	c.RepoTree = NewRepoTreeClient(conn)
	c.Search = NewSearchClient(conn)
	c.Units = NewUnitsClient(conn)
	c.Users = NewUsersClient(conn)
	c.Defs = NewDefsClient(conn)
	c.Markdown = NewMarkdownClient(conn)

	c.UserAgent = userAgent

	return c
}

// NewClient returns a new Sourcegraph API client. If httpClient is nil,
// http.DefaultClient is used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		cloned := *http.DefaultClient
		httpClient = &cloned
	}

	// TODO(sqs!nodb-ctx): this needs to be filled in to use the
	// grpc-gateway client impls, if we go that route

	c := new(Client)
	c.httpClient = httpClient

	c.BaseURL = &url.URL{Scheme: "https", Host: "sourcegraph.com", Path: "/api/"}

	c.UserAgent = userAgent

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
func newResponse(r *http.Response) *HTTPResponse {
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
func (c *Client) Do(req *http.Request, v interface{}) (*HTTPResponse, error) {
	var resp *HTTPResponse
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
