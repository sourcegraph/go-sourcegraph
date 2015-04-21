package sourcegraph

// SearchService communicates with the search-related endpoints in
// the Sourcegraph API.
type SearchService interface {
	// Search searches the full index.
	Search(opt *SearchOptions) (*SearchResults, error)

	// Complete completes the token at the RawQuery's InsertionPoint.
	Complete(q RawQuery) (*Completions, error)

	// Suggest suggests queries given an existing query. It can be
	// called with an empty query to get example queries that pertain
	// to the current user's repositories, orgs, etc.
	Suggest(q RawQuery) ([]*Suggestion, error)
}

// Empty is whether there are no search results for any result type.
func (r *SearchResults) Empty() bool {
	return len(r.Defs) == 0 && len(r.People) == 0 && len(r.Repos) == 0 && len(r.Tree) == 0
}
