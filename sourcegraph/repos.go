package sourcegraph

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/go-nnz/nnz"

	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"strings"
)

// ReposService communicates with the repository-related endpoints in the
// Sourcegraph API.
type ReposService interface {
	// Get fetches a repository.
	Get(repo RepoSpec, opt *RepoGetOptions) (*Repo, error)

	// List repositories.
	List(opt *RepoListOptions) ([]*Repo, error)

	// Create adds a repository.
	Create(newRepo *Repo) (*Repo, error)

	// GetReadme fetches the formatted README file for a repository.
	GetReadme(repo RepoRevSpec) (*vcsclient.TreeEntry, error)
}

// Repo represents a source code repository.
type Repo struct {
	// URI is a normalized identifier for this repository based on its primary
	// clone URL. E.g., "github.com/user/repo".
	URI string

	// URIAlias is another URI that, if accessed, will redirect to
	// this repository's primary URI. It's used, for example, to
	// redirect from GitHub repos to their canonical URI (such as Go
	// repos on gopkg.in).
	URIAlias nnz.String `db:"uri_alias"`

	// Name is the base name (the final path component) of the repository,
	// typically the name of the directory that the repository would be cloned
	// into. (For example, for git://example.com/foo.git, the name is "foo".)
	Name string

	// OwnerUserID is the account that owns this repository.
	OwnerUserID int `db:"owner_user_id"`

	// OwnerGitHubUserID is the GitHub user ID of this repository's owner, if this
	// is a GitHub repository.
	OwnerGitHubUserID nnz.Int `db:"owner_github_user_id" json:",omitempty"`

	// Description is a brief description of the repository.
	Description string `json:",omitempty"`

	// VCS is the short name of the VCS system that this repository uses: "git"
	// or "hg".
	VCS string `db:"vcs"`

	// HTTPCloneURL is the HTTPS clone URL of the repository (or the
	// HTTP clone URL, if no HTTPS clone URL is available).
	HTTPCloneURL string `db:"http_clone_url"`

	// SSHCloneURL is the SSH clone URL if the repository, if any.
	SSHCloneURL nnz.String `db:"ssh_clone_url"`

	// HomepageURL is the URL to the repository's homepage, if any.
	HomepageURL nnz.String `db:"homepage_url"`

	// DefaultBranch is the default VCS branch used (typically "master" for git
	// repositories and "default" for hg repositories).
	DefaultBranch string `db:"default_branch"`

	// Language is the primary programming language used in this repository.
	Language string

	// GitHubStars is the number of stargazers this repository has on GitHub (or
	// 0 if it is not a GitHub repository).
	GitHubStars int `db:"github_stars"`

	// GitHubID is the GitHub ID of this repository. If a GitHub repository is
	// renamed, the ID remains the same and should be used to resolve across the
	// name change.
	GitHubID nnz.Int `db:"github_id" json:",omitempty"`

	// Disabled is whether this repo should not be downloaded and processed by the worker.
	Disabled bool `json:",omitempty"`

	// Deprecated repositories are labeled as such and hidden from global search results.
	Deprecated bool

	// Fork is whether this repository is a fork.
	Fork bool

	// Mirror is whether this repository is a mirror.
	Mirror bool

	// Private is whether this repository is private.
	Private bool

	// CreatedAt is when this repository was created. If it represents
	// an externally hosted (e.g., GitHub) repository, the creation
	// date is when it was created at that origin.
	CreatedAt time.Time `db:"created_at"`

	// UpdatedAt is when this repository's metadata was last updated
	// (on its origin if it's an externally hosted repository).
	UpdatedAt time.Time `db:"updated_at"`

	// PushedAt is when this repository's was last (VCS-)pushed to.
	PushedAt time.Time `db:"pushed_at"`

	// Permissions describes the permissions that the current user (or
	// anonymous users, if there is no current user) is granted to
	// this repository.
	Permissions *RepoPermissions `db:"-" json:",omitempty"`
}

// IsGitHubRepo returns true iff this repository is hosted on GitHub.
func (r *Repo) IsGitHubRepo() bool { return r.GitHubID != 0 }

// Returns the repository's canonical clone URL
func (r *Repo) CloneURL() *url.URL {
	var cloneURL string
	if r.HTTPCloneURL != "" {
		cloneURL = r.HTTPCloneURL
	} else if r.SSHCloneURL != "" {
		cloneURL = string(r.SSHCloneURL)
	} else {
		cloneURL = r.URI
	}
	u, _ := url.Parse(cloneURL)
	return u
}

// GitHubHTMLURL returns URL to the GitHub HTML page (e.g.,
// https://github.com/foo/bar, not a clone URL) for this repo, if it's
// a GitHub repo. Otherwise it returns the empty string.
func (r *Repo) GitHubHTMLURL() string {
	var ghuri string
	if IsGitHubRepoURI(r.URI) {
		ghuri = r.URI
	} else if IsGitHubRepoURI(string(r.URIAlias)) {
		ghuri = string(r.URIAlias)
	}
	if ghuri == "" {
		return ""
	}
	return (&url.URL{Scheme: "https", Host: "github.com", Path: "/" + strings.TrimPrefix(ghuri, githubRepoURIPrefix)}).String()
}

// RepoSpec specifies a repository.
type RepoSpec struct {
	URI string
}

// IsZero reports whether s.URI is the zero value.
func (s RepoSpec) IsZero() bool { return s.URI == "" }

// PathComponent returns the URL path component that specifies the
// repository.
func (s RepoSpec) PathComponent() string {
	if s.IsZero() {
		panic("IsZero")
	}
	return s.URI
}

// RouteVars returns route variables for constructing repository
// routes.
func (s RepoSpec) RouteVars() map[string]string {
	return map[string]string{"RepoSpec": s.PathComponent()}
}

// ParseRepoSpec parses a string generated by
// (*RepoSpec).PathComponent() and returns the equivalent
// RepoSpec struct.
func ParseRepoSpec(pathComponent string) (RepoSpec, error) {
	if pathComponent == "" {
		return RepoSpec{}, errors.New("empty repository spec")
	}
	return RepoSpec{URI: pathComponent}, nil
}

// UnmarshalRepoSpec marshals a map containing route variables
// generated by (*RepoSpec).RouteVars() and returns the
// equivalent RepoSpec struct.
func UnmarshalRepoSpec(routeVars map[string]string) (RepoSpec, error) {
	return ParseRepoSpec(routeVars["RepoSpec"])
}

// RepoRevSpec specifies a repository at a specific commit (or
// revision specifier, such as a branch, which is resolved on the
// server side to a specific commit).
//
// Filling in CommitID is an optional optimization. It avoids the need
// for another resolution of Rev. If CommitID is filled in, the "Rev"
// route variable becomes "Rev===CommitID" (e.g.,
// "master===af4cd6"). Handlers can parse this string to retrieve the
// pre-resolved commit ID (e.g., "af4cd6") and still return data that
// constructs URLs using the unresolved revspec (e.g., "master").
//
// Why is it important/useful to pass the resolved commit ID instead
// of just using a revspec everywhere? Consider this case. Your
// application wants to make a bunch of requests for resources
// relating to "master"; for example, it wants to retrieve a source
// file foo.go at master and all of the definitions and references
// contained in the file. This may consist of dozens of API calls. If
// each API call specified just "master", there would be 2 problems:
// (1) each API call would have to re-resolve "master" to its actual
// commit ID, which takes a lot of extra work; and (2) if the "master"
// ref changed during the API calls (if someone pushed in the middle
// of the API call, for example), then your application would receive
// data from 2 different commits. The solution is for your application
// to resolve the revspec once and pass both the original revspec and
// the resolved commit ID in all API calls it makes.
//
// And why do we want to preserve the unresolved revspec? In this
// case, your app wants to let the user continue browsing "master". If
// the API data all referred to a specific commit ID, then the user
// would cease browsing master the next time she clicked a link on
// your app. Preserving the revspec gives the user a choice whether to
// use the absolute commit ID or the revspec (similar to how GitHub
// lets you canonicalize a URL with 'y' but does not default to using
// the canonical URL).
type RepoRevSpec struct {
	RepoSpec        // repository specifier
	Rev      string // the abstract/unresolved revspec, such as a branch name or abbreviated commit ID
	CommitID string // the full commit ID that Rev resolves to
}

const repoRevSpecCommitSep = "==="

// RouteVars returns route variables for constructing routes to a
// repository commit.
func (s RepoRevSpec) RouteVars() map[string]string {
	m := s.RepoSpec.RouteVars()
	m["Rev"] = s.RevPathComponent()
	return m
}

// RevPathComponent encodes the revision and commit ID for use in a
// URL path. If CommitID is set, the path component is
// "Rev===CommitID"; otherwise, it is just "Rev". See the docstring
// for RepoRevSpec for an explanation why.
func (s RepoRevSpec) RevPathComponent() string {
	if s.Rev == "" && s.CommitID != "" {
		panic("invalid empty Rev but non-empty CommitID (" + s.CommitID + ")")
	}
	if s.CommitID != "" {
		return s.Rev + repoRevSpecCommitSep + s.CommitID
	}
	return s.Rev
}

// UnmarshalRepoRevSpec marshals a map containing route variables
// generated by (*RepoRevSpec).RouteVars() and returns the equivalent
// RepoRevSpec struct.
func UnmarshalRepoRevSpec(routeVars map[string]string) (RepoRevSpec, error) {
	repoSpec, err := UnmarshalRepoSpec(routeVars)
	if err != nil {
		return RepoRevSpec{}, err
	}

	repoRevSpec := RepoRevSpec{RepoSpec: repoSpec}
	revStr := routeVars["Rev"]
	if i := strings.Index(revStr, repoRevSpecCommitSep); i == -1 {
		repoRevSpec.Rev = revStr
	} else {
		repoRevSpec.Rev = revStr[:i]
		repoRevSpec.CommitID = revStr[i+len(repoRevSpecCommitSep):]
	}

	if repoRevSpec.Rev == "" && repoRevSpec.CommitID != "" {
		return RepoRevSpec{}, fmt.Errorf("invalid empty Rev but non-empty CommitID (%q)", repoRevSpec.CommitID)
	}

	return repoRevSpec, nil
}

// RepoGetOptions specifies options for getting a repository.
type RepoGetOptions struct{}

type RepoListOptions struct {
	Name string `url:",omitempty" json:",omitempty"`

	// Specifies a search query for repositories. If specified, then the Sort and Direction options are ignored
	Query string `url:",omitempty" json:",omitempty"`

	URIs []string `url:",comma,omitempty" json:",omitempty"`

	BuiltOnly bool `url:",omitempty" json:",omitempty"`

	Sort      string `url:",omitempty" json:",omitempty"`
	Direction string `url:",omitempty" json:",omitempty"`

	NoFork bool `url:",omitempty" json:",omitempty"`

	Type string `url:",omitempty" json:",omitempty"` // "public" or "private" (empty default means "all")

	State string `url:",omitempty" json:",omitempty"` // "enabled" or "disabled" (empty default means return "all")

	Owner string `url:",omitempty" json:",omitempty"`

	ListOptions
}
