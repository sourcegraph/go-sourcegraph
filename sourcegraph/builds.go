package sourcegraph

import (
	"errors"
	"fmt"

	"strconv"
)

func (s *BuildSpec) RouteVars() map[string]string {
	m := s.RepoRev.RouteVars()
	m["BID"] = fmt.Sprintf("%d", s.BID)
	return m
}

func (s *TaskSpec) RouteVars() map[string]string {
	v := s.BuildSpec.RouteVars()
	v["TaskID"] = fmt.Sprintf("%d", s.TaskID)
	return v
}

func (b *Build) Spec() BuildSpec {
	return BuildSpec{
		RepoRev: RepoRevSpec{
			RepoSpec: RepoSpec{URI: b.Repo},
			CommitID: b.CommitID,
		},
		BID: b.BID,
	}
}

// IDString returns a succinct string that uniquely identifies this build.
func (b BuildSpec) IDString() string { return buildIDString(b.BID) }

func buildIDString(bid int64) string { return "B" + strconv.FormatInt(bid, 36) }

// Build task ops.
const (
	ImportTaskOp = "import"
)

func (t *BuildTask) Spec() TaskSpec {
	return TaskSpec{
		BuildSpec: BuildSpec{
			BID: t.BID,
			RepoRev: RepoRevSpec{
				RepoSpec: RepoSpec{URI: t.Repo},
				CommitID: t.CommitID,
			},
		},
		TaskID: t.TaskID,
	}
}

// Update sets each field on t that is non-nil in update.
func (t *BuildTask) Update(update TaskUpdate) {
	if update.StartedAt != nil {
		tmp := *update.StartedAt // Copy
		t.StartedAt = &tmp
	}
	if update.EndedAt != nil {
		tmp := *update.EndedAt // Copy
		t.EndedAt = &tmp
	}
	// SAMER: figure out this logic.
	// if update.Success != nil {
	// 	t.Success = *update.Success
	// }
	// if update.Failure != nil {
	// 	t.Failure = *update.Failure
	// }
}

// IDString returns a succinct string that uniquely identifies this build task.
func (t TaskSpec) IDString() string {
	return buildIDString(t.BID) + "-T" + strconv.FormatInt(t.TaskID, 36)
}

var ErrBuildNotFound = errors.New("build not found")
