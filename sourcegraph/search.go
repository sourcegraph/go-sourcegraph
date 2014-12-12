package sourcegraph

import "sourcegraph.com/sourcegraph/go-sourcegraph/router"

// SearchService communicates with the search-related endpoints in
// the Sourcegraph API.
type SearchService interface {
	// Search searches the full index.
	Search(opt *SearchOptions) (*SearchResults, Response, error)
}

type SearchResults struct {
	Defs   []*Def
	People []*Person
	Repos  []*Repo
}

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
	url, err := s.client.url(router.Search, nil, opt)
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

type MockSearchService struct {
	Search_ func(opt *SearchOptions) (*SearchResults, Response, error)
}

var _ SearchService = MockSearchService{}

func (s MockSearchService) Search(opt *SearchOptions) (*SearchResults, Response, error) {
	return s.Search_(opt)
}
