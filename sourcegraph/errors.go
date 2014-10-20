package sourcegraph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

// An ErrorResponse reports errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:",omitempty"` // HTTP response that caused this error
	Message  string         // error message
	// Kind is the type of error recieved.
	Kind ErrorResponseKind
}

type ErrorResponseKind string

const (
	DefError ErrorResponseKind = "deferror"
)

func parseErrorKind(errorMessage string) ErrorResponseKind {
	if strings.Contains(errorMessage, graph.ErrDefNotExist.Error()) {
		return DefError
	}
	return ""
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

func (r *ErrorResponse) HTTPStatusCode() int { return r.Response.StatusCode }

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.  API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.  Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	errorResponse.Kind = parseErrorKind(errorResponse.Message)
	return errorResponse
}

func IsHTTPErrorCode(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	type httpError interface {
		Error() string
		HTTPStatusCode() int
	}
	if httpErr, ok := err.(httpError); ok {
		return statusCode == httpErr.HTTPStatusCode()
	}
	return false
}
