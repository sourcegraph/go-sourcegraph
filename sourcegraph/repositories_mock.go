package sourcegraph

import (
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
)

type MockReposService struct {
	Get_               func(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error)
	CreateStatus_      func(spec RepoRevSpec, st RepoStatus) (*RepoStatus, Response, error)
	GetCombinedStatus_ func(spec RepoRevSpec) (*CombinedStatus, Response, error)
	GetSettings_       func(repo RepoSpec) (*RepoSettings, Response, error)
	UpdateSettings_    func(repo RepoSpec, settings RepoSettings) (Response, error)
	RefreshVCSData_    func(repo RepoSpec) (Response, error)
	Create_            func(newRepo *Repo) (*Repo, Response, error)
	GetReadme_         func(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error)
	List_              func(opt *RepoListOptions) ([]*Repo, Response, error)
	ListCommits_       func(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error)
	GetCommit_         func(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error)
	ListBranches_      func(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error)
	ListTags_          func(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error)
	ListBadges_        func(repo RepoSpec) ([]*Badge, Response, error)
	ListCounters_      func(repo RepoSpec) ([]*Counter, Response, error)
}

func (s *MockReposService) Get(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error) {
	return s.Get_(repo, opt)
}

func (s *MockReposService) CreateStatus(spec RepoRevSpec, st RepoStatus) (*RepoStatus, Response, error) {
	return s.CreateStatus_(spec, st)
}

func (s *MockReposService) GetCombinedStatus(spec RepoRevSpec) (*CombinedStatus, Response, error) {
	return s.GetCombinedStatus_(spec)
}

func (s *MockReposService) GetSettings(repo RepoSpec) (*RepoSettings, Response, error) {
	return s.GetSettings_(repo)
}

func (s *MockReposService) UpdateSettings(repo RepoSpec, settings RepoSettings) (Response, error) {
	return s.UpdateSettings_(repo, settings)
}

func (s *MockReposService) RefreshVCSData(repo RepoSpec) (Response, error) {
	return s.RefreshVCSData_(repo)
}

func (s *MockReposService) Create(newRepo *Repo) (*Repo, Response, error) { return s.Create_(newRepo) }

func (s *MockReposService) GetReadme(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error) {
	return s.GetReadme_(repo)
}

func (s *MockReposService) List(opt *RepoListOptions) ([]*Repo, Response, error) { return s.List_(opt) }

func (s *MockReposService) ListCommits(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error) {
	return s.ListCommits_(repo, opt)
}

func (s *MockReposService) GetCommit(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error) {
	return s.GetCommit_(rev, opt)
}

func (s *MockReposService) ListBranches(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error) {
	return s.ListBranches_(repo, opt)
}

func (s *MockReposService) ListTags(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error) {
	return s.ListTags_(repo, opt)
}

func (s *MockReposService) ListBadges(repo RepoSpec) ([]*Badge, Response, error) {
	return s.ListBadges_(repo)
}

func (s *MockReposService) ListCounters(repo RepoSpec) ([]*Counter, Response, error) {
	return s.ListCounters_(repo)
}
