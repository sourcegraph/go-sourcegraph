package sourcegraph

// TODO(sqs!nodb-ctx): can't protobuf-ify the BuildDataService because
// it returns a Go interface, not data.

// BuildDataFileSpec specifies a new or existing build data file in a
// repository.
type BuildDataFileSpec struct {
	RepoRev RepoRevSpec
	Path    string
}

// RouteVars returns route variables used to construct URLs to a build
// data file.
func (s *BuildDataFileSpec) RouteVars() map[string]string {
	m := s.RepoRev.RouteVars()
	m["Path"] = s.Path
	return m
}
