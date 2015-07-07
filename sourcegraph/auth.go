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

// UserSpec returns a UserSpec that refers to the user identified by
// a. If a.UID == 0, nil is returned.
func (a AuthInfo) UserSpec() *UserSpec {
	if a.UID == 0 {
		return nil
	}
	return &UserSpec{UID: a.UID, Domain: a.Domain}
}
