package sourcegraph

type RepoSettingsService interface {
	// GetSettings fetches a repository's configuration settings.
	GetSettings(repo RepoSpec) (*RepoSettings, error)

	// UpdateSettings updates a repository's configuration settings.
	UpdateSettings(repo RepoSpec, settings RepoSettings) error
}

// RepoSettings describes a repository's configuration settings.
type RepoSettings struct {
	// Enabled is whether this repository has been enabled for use on
	// Sourcegraph by a repository owner or a site admin.
	Enabled *bool `db:"enabled" json:",omitempty"`

	// BuildPushes is whether head commits on newly pushed branches
	// should be automatically built.
	BuildPushes *bool `db:"build_pushes" json:",omitempty"`

	// ExternalCommitStatuses is whether the build status
	// (pending/failure/success) of each commit should be published to
	// GitHub using the repo status API
	// (https://developer.github.com/v3/repos/statuses/).
	//
	// This behavior is also subject to the
	// UnsuccessfulExternalCommitStatuses setting value.
	ExternalCommitStatuses *bool `db:"external_commit_statuses" json:",omitempty"`

	// UnsuccessfulExternalCommitStatuses, if true, indicates that
	// pending/failure commit statuses should be published to
	// GitHub. If false (default), only successful commit status are
	// published. The default of false avoids bothersome warning
	// messages and UI pollution in case the Sourcegraph build
	// fails. Until our builds are highly reliable, a Sourcegraph
	// build failure is not necessarily an indication of a problem
	// with the repository.
	//
	// This setting is only meaningful if ExternalCommitStatuses is
	// true.
	UnsuccessfulExternalCommitStatuses *bool `db:"unsuccessful_external_commit_statuses" json:",omitempty"`

	// UseSSHPrivateKey is whether Sourcegraph should clone and update
	// the repository using an SSH key, and whether it should copy the
	// corresponding public key to the repository's origin host as an
	// authorized key. It is only necessary for private repositories
	// and for write operations on public repositories.
	UseSSHPrivateKey *bool `db:"use_ssh_private_key" json:",omitempty"`

	// LastAdminUID is the UID of the last user to modify this repo's
	// settings. When Sourcegraph needs to perform actions on GitHub
	// repos that require OAuth authorization outside of an authorized
	// HTTP request (e.g., during builds or asynchronous operations),
	// it consults the repo's LastAdminUID to determine whose identity
	// it should assume to perform the operation.
	//
	// If the LastAdminUID refers to a user who no longer has
	// permissions to perform the action, GitHub will refuse to
	// perform the operation. In that case, another admin of the
	// repository needs to update the settings so that she will become
	// the new LastAdminUID.
	LastAdminUID *int `db:"admin_uid" json:",omitempty"`
}
