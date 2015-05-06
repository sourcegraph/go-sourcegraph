package mock

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func (s *ReposServer) MockGet(t *testing.T, wantRepo string) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.URI != wantRepo {
			t.Errorf("got repo %q, want %q", repo.URI, wantRepo)
			return nil, sourcegraph.ErrNotExist
		}
		return &sourcegraph.Repo{URI: repo.URI}, nil
	}
	return
}

func (s *ReposServer) MockGet_Return(t *testing.T, returns *sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.URI != returns.URI {
			t.Errorf("got repo %q, want %q", repo.URI, returns.URI)
			return nil, sourcegraph.ErrNotExist
		}
		return returns, nil
	}
	return
}

func (s *ReposServer) MockList(t *testing.T, wantRepos ...string) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
		*called = true
		repos := make([]*sourcegraph.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &sourcegraph.Repo{URI: repo}
		}
		return &sourcegraph.RepoList{Repos: repos}, nil
	}
	return
}

func (s *ReposServer) MockListCommits(t *testing.T, wantCommitIDs ...vcs.CommitID) (called *bool) {
	called = new(bool)
	s.ListCommits_ = func(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
		*called = true
		commits := make([]*vcs.Commit, len(wantCommitIDs))
		for i, commit := range wantCommitIDs {
			commits[i] = &vcs.Commit{ID: commit}
		}
		return &sourcegraph.CommitList{Commits: commits}, nil
	}
	return
}

func (s *ReposServer) MockGetCommit_ByID_NoCheck(t *testing.T, commitID vcs.CommitID) (called *bool) {
	called = new(bool)
	s.GetCommit_ = func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		*called = true
		return &vcs.Commit{ID: commitID}, nil
	}
	return
}

func (s *ReposServer) MockGetCommit_Return_NoCheck(t *testing.T, commit *vcs.Commit) (called *bool) {
	called = new(bool)
	s.GetCommit_ = func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		*called = true
		return commit, nil
	}
	return
}
