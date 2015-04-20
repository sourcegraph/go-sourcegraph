package sourcegraph

import "errors"

var (
	// ErrRepoBlocked occurs when an operation is called on a repository
	// that is blocked.
	ErrRepoBlocked = errors.New("repo is blocked")

	ErrNoDefaultBranch = errors.New("repo has no default branch")
)
