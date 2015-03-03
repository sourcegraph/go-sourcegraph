package sourcegraph

import "sourcegraph.com/sourcegraph/rwvfs"

type MockBuildDataService struct {
	FileSystem_ func(repo RepoRevSpec) (rwvfs.FileSystem, error)
}

func (s MockBuildDataService) FileSystem(repo RepoRevSpec) (rwvfs.FileSystem, error) {
	return s.FileSystem_(repo)
}
