package sourcegraph

import "net/http"

// An AuthenticationError occurs when incorrect authentication
// credentials are provided.
type AuthenticationError struct {
	Kind string // "cookie" (session cookie) or "key" (API key in Authorization header)
	Err  error
}

func (e *AuthenticationError) Error() string { return "authentication failure: " + e.Err.Error() }

// HTTPStatusCode implements handlerutil.HTTPError.
func (e *AuthenticationError) HTTPStatusCode() int { return http.StatusForbidden }
