package sourcegraph

import (
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// SearchService communicates with the search-related endpoints in
// the Sourcegraph API.
type SearchService interface {
	// Search searches the full index.
	Search(opt *SearchOptions) (*SearchResults, Response, error)

	// Complete completes the token at the RawQuery's InsertionPoint.
	Complete(q RawQuery) (*Completions, Response, error)
}

type SearchResults struct {
	Defs   []*Def    `json:",omitempty"`
	People []*Person `json:",omitempty"`
	Repos  []*Repo   `json:",omitempty"`

	// RawQuery is the raw query passed to search.
	RawQuery RawQuery

	// Tokens are the unresolved tokens.
	Tokens Tokens `json:",omitempty"`

	// Plan is the query plan used to fetch the results.
	Plan *Plan `json:",omitempty"`

	// ResolvedTokens holds the resolved tokens from the original query
	// string.
	ResolvedTokens Tokens

	ResolveErrors []TokenError `json:",omitempty"`

	// Tips are helpful tips for the user about their query. They are
	// not errors per se, but they use the TokenError type because it
	// allows us to associate a message with a particular token (and
	// JSON de/serialize that).
	Tips []TokenError `json:",omitempty"`

	// Canceled is true if the query was canceled. More information
	// about how to correct the issue can be found in the
	// ResolveErrors and Tips.
	Canceled bool
}

// Empty is whether there are no search results for any result type.
func (r *SearchResults) Empty() bool {
	return len(r.Defs) == 0 && len(r.People) == 0 && len(r.Repos) == 0
}

// searchService implements SearchService.
type searchService struct {
	client *Client
}

var _ SearchService = &searchService{}

type SearchOptions struct {
	Query string `url:"q" schema:"q"`

	Defs   bool
	Repos  bool
	People bool

	ListOptions
}

func (s *searchService) Search(opt *SearchOptions) (*SearchResults, Response, error) {
	url, err := s.client.URL(router.Search, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var results *SearchResults
	resp, err := s.client.Do(req, &results)
	if err != nil {
		return nil, resp, err
	}

	return results, resp, nil
}

// Completions holds search query completions.
type Completions struct {
	// TokenCompletions are suggested completions for the token at the
	// raw query's InsertionPoint.
	TokenCompletions Tokens

	// ResolvedTokens is the resolution of the original query's tokens
	// used to produce the completions. It is useful for debugging.
	ResolvedTokens Tokens

	ResolveErrors   []TokenError `json:",omitempty"`
	ResolutionFatal bool         `json:",omitempty"`
}

func (s *searchService) Complete(q RawQuery) (*Completions, Response, error) {
	url, err := s.client.URL(router.SearchComplete, nil, q)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var comps *Completions
	resp, err := s.client.Do(req, &comps)
	if err != nil {
		return nil, resp, err
	}

	return comps, resp, nil
}

type MockSearchService struct {
	Search_   func(opt *SearchOptions) (*SearchResults, Response, error)
	Complete_ func(q RawQuery) (*Completions, Response, error)
}

var _ SearchService = MockSearchService{}

func (s MockSearchService) Search(opt *SearchOptions) (*SearchResults, Response, error) {
	return s.Search_(opt)
}

func (s MockSearchService) Complete(q RawQuery) (*Completions, Response, error) {
	return s.Complete_(q)
}
