package sourcegraph

type MockSearchService struct {
	Search_ func(opt *SearchOptions) (*SearchResults, Response, error)
}

func (s *MockSearchService) Search(opt *SearchOptions) (*SearchResults, Response, error) {
	return s.Search_(opt)
}
