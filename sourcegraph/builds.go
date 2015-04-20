package sourcegraph

import (
	"errors"
	"fmt"

	"strconv"
)

func (s *BuildSpec) RouteVars() map[string]string {
	m := s.Repo.RouteVars()
	m["BID"] = fmt.Sprintf("%d", s.BID)
	return m
}

func (s *TaskSpec) RouteVars() map[string]string {
	v := s.BuildSpec.RouteVars()
	v["TaskID"] = fmt.Sprintf("%d", s.TaskID)
	return v
}

func (b *Build) Spec() BuildSpec { return BuildSpec{Repo: RepoSpec{URI: b.Repo}, BID: b.BID} }

// IDString returns a succinct string that uniquely identifies this build.
func (b BuildSpec) IDString() string { return buildIDString(b.BID) }

func buildIDString(bid int64) string { return "B" + strconv.FormatInt(bid, 36) }

// Build task ops.
const (
	ImportTaskOp = "import"
)

func (t *BuildTask) Spec() TaskSpec {
	return TaskSpec{BuildSpec: BuildSpec{Repo: RepoSpec{URI: t.Repo}, BID: t.BID}, TaskID: t.TaskID}
}

// IDString returns a succinct string that uniquely identifies this build task.
func (t TaskSpec) IDString() string {
	return buildIDString(t.BID) + "-T" + strconv.FormatInt(t.TaskID, 36)
}

var ErrBuildNotFound = errors.New("build not found")
