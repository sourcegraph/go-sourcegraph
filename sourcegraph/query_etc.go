package sourcegraph

import (
	"encoding/json"
	"fmt"
)

// A Plan is a query plan that fetches the data necessary to satisfy
// (and provide autocomplete suggestions for) a query.
type Plan struct {
	Repos *RepoListOptions
	Defs  *DefListOptions
	Users *UsersListOptions
}

func (p *Plan) String() string {
	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}

// ResolveErrors occurs when query validation fails. It is returned
// by Resolve and can hold multiple errors.
type ResolveErrors []ResolveError

func (e ResolveErrors) Error() string {
	return fmt.Sprintf("%d resolution errors: %v", len(e), ([]ResolveError)(e))
}

// Fatal is true if the query may not be processed or executed further
// due to unrecoverable query resolution errors.
func (e ResolveErrors) Fatal() bool { return true }

// A ResolveError is returned inside a ResolveErrors list by
// Resolve and occurs when query validation fails.
type ResolveError struct {
	Token  Token  // the token that caused the error
	Reason string // the public, user-readable error message to display
}

func (e ResolveError) Error() string { return fmt.Sprintf("%s (%v)", e.Reason, e.Token) }

type jsonResolveError struct {
	Token  jsonToken
	Reason string
}

func (e ResolveError) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonResolveError{jsonToken{e.Token}, e.Reason})
}

func (e *ResolveError) UnmarshalJSON(b []byte) error {
	var jv jsonResolveError
	if err := json.Unmarshal(b, &jv); err != nil {
		return err
	}
	*e = ResolveError{jv.Token.Token, jv.Reason}
	return nil
}

// IsResolveError returns true if err is a ResolveErrors list or a
// single ResolveError.
func IsResolveError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(ResolveErrors); ok {
		return true
	}
	if _, ok := err.(ResolveError); ok {
		return true
	}
	return false
}
