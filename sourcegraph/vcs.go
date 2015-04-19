package sourcegraph

import "sourcegraph.com/sourcegraph/go-vcs/vcs"

type VCSOpener interface {
	OpenVCS(RepoSpec) (vcs.Repository, error)
}
