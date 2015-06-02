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
}

func IsDefError(err error) bool {
	return err != nil && strings.Contains(err.Error(), graph.ErrDefNotExist.Error())
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

// InvalidOptionsError indicates that the provided XxxOptions
// (RepoListOptions, DefGetOptions, etc.) was invalid.
type InvalidOptionsError struct{ Reason string }

func (e *InvalidOptionsError) Error() string { return e.Reason }

func (e *InvalidOptionsError) HTTPStatusCode() int { return http.StatusBadRequest }

// InvalidSpecError indicates that the provided XxxSpec (RepoSpec,
// UserSpec, etc.) was invalid.
type InvalidSpecError struct{ Reason string }

func (e *InvalidSpecError) Error() string { return e.Reason }

func (e *InvalidSpecError) HTTPStatusCode() int { return http.StatusBadRequest }

// NotImplementedError indicates that a not-yet-implemented method was
// called.
type NotImplementedError struct{ What string }

func (e *NotImplementedError) Error() string { return e.What + " is not implemented" }

func (e *NotImplementedError) HTTPStatusCode() int { return http.StatusNotFound }
