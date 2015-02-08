package sourcegraph

type MockRepoTreeService struct {
	Get_ func(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error)
}

func (s MockRepoTreeService) Get(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error) {
	return s.Get_(entry, opt)
}
