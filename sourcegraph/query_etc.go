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

// A ResolveError is returned inside a ResolveErrors list by
// Resolve and occurs when query validation fails.
type ResolveError struct {
	// Index is the 1-indexed index of the token that caused the error
	// (0 means not associated with any particular token).
	//
	// NOTE: Index is 1-indexed (not 0-indexed) because some
	// ResolveErrors don't pertain to a token, and it's misleading if
	// the Index in the JSON is 0 (which could mean that it pertains
	// to the 1st token if index was 0-indexed).
	Index int `json:",omitempty"`

	Token  Token  `json:",omitempty"` // the token that caused the error
	Reason string // the public, user-readable error message to display
}

func (e ResolveError) Error() string { return fmt.Sprintf("%s (%v)", e.Reason, e.Token) }

type jsonResolveError struct {
	Index  int       `json:",omitempty"`
	Token  jsonToken `json:",omitempty"`
	Reason string
}

func (e ResolveError) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonResolveError{e.Index, jsonToken{e.Token}, e.Reason})
}

func (e *ResolveError) UnmarshalJSON(b []byte) error {
	var jv jsonResolveError
	if err := json.Unmarshal(b, &jv); err != nil {
		return err
	}
	*e = ResolveError{jv.Index, jv.Token.Token, jv.Reason}
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
