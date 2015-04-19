package sourcegraph

import "time"

type RepoStatusService interface {
	// Create creates a repository status for the given commit.
	Create(spec RepoRevSpec, st RepoStatus) (*RepoStatus, error)

	// GetCombined fetches the combined repository status for
	// the given commit.
	GetCombined(spec RepoRevSpec) (*CombinedStatus, error)
}

// RepoStatus is the status of the repository at a specific rev (in a
// single context).
type RepoStatus struct {
	// CommitID is the full commit ID of the commit this status
	// describes.
	CommitID string

	// State is the current status of the repository. Possible values are:
	// pending, success, error, or failure.
	State string

	// TargetURL is the URL of the page representing this status. It will be
	// linked from the UI to allow users to see the source of the status.
	TargetURL string

	// Description is a short, high-level summary of the status.
	Description string

	// A string label to differentiate this status from the statuses of other systems.
	Context string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CombinedStatus is the combined status (i.e., incorporating statuses
// from all contexts) of the repository at a specific rev.
type CombinedStatus struct {
	// CommitID is the full commit ID of the commit this status
	// describes.
	CommitID string

	// State is the combined status of the repository. Possible values are:
	// failture, pending, or success.
	State string

	// Statuses are the statuses for each context.
	Statuses []*RepoStatus
}
