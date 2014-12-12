package sourcegraph

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/sourcegraph/go-nnz/nnz"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// ReposService communicates with the repository-related endpoints in the
// Sourcegraph API.
type ReposService interface {
	// Get fetches a repository.
	Get(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error)

	// GetStats gets statistics about a repository at a specific
	// commit. Some statistics are per-commit and some are global to
	// the repository. If you only care about global repository
	// statistics, pass an empty Rev to the RepoRevSpec (which will be
	// resolved to the repository's default branch).
	GetStats(repo RepoRevSpec) (RepoStats, Response, error)

	// GetOrCreate fetches a repository using Get. If no such repository exists
	// with the URI, and the URI refers to a recognized repository host (such as
	// github.com), the repository's information is fetched from the external
	// host and the repository is created.
	GetOrCreate(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error)

	// GetSettings fetches a repository's configuration settings.
	GetSettings(repo RepoSpec) (*RepoSettings, Response, error)

	// UpdateSettings updates a repository's configuration settings.
	UpdateSettings(repo RepoSpec, settings RepoSettings) (Response, error)

	// RefreshProfile updates the repository metadata for a repository, fetching
	// it from an external host if the host is recognized (such as GitHub).
	//
	// This operation is performed asynchronously on the server side (after
	// receiving the request) and the API currently has no way of notifying
	// callers when the operation completes.
	RefreshProfile(repo RepoSpec) (Response, error)

	// RefreshVCSData updates the repository VCS (git/hg) data, fetching all new
	// commits, branches, tags, and blobs.
	//
	// This operation is performed asynchronously on the server side (after
	// receiving the request) and the API currently has no way of notifying
	// callers when the operation completes.
	RefreshVCSData(repo RepoSpec) (Response, error)

	// ComputeStats updates the statistics about a repository.
	//
	// This operation is performed asynchronously on the server side (after
	// receiving the request) and the API currently has no way of notifying
	// callers when the operation completes.
	ComputeStats(repo RepoRevSpec) (Response, error)

	// GetBuild gets the build for a specific revspec. It returns
	// additional information about the build, such as whether it is
	// exactly up-to-date with the revspec or a few commits behind the
	// revspec. The opt param controls what is returned in this case.
	GetBuild(repo RepoRevSpec, opt *RepoGetBuildOptions) (*RepoBuildInfo, Response, error)

	// Create adds the repository at cloneURL, filling in all information about
	// the repository that can be inferred from the URL (or, for GitHub
	// repositories, fetched from the GitHub API). If a repository with the
	// specified clone URL, or the same URI, already exists, it is returned.
	Create(newRepoSpec NewRepoSpec) (*Repo, Response, error)

	// GetReadme fetches the formatted README file for a repository.
	GetReadme(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error)

	// List repositories.
	List(opt *RepoListOptions) ([]*Repo, Response, error)

	// List commits.
	ListCommits(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error)

	// GetCommit gets a commit.
	GetCommit(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error)

	// ListBranches lists a repository's branches.
	ListBranches(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error)

	// ListTags lists a repository's tags.
	ListTags(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error)

	// ListBadges lists the available badges for repo.
	ListBadges(repo RepoSpec) ([]*Badge, Response, error)

	// ListCounters lists the available counters for repo.
	ListCounters(repo RepoSpec) ([]*Counter, Response, error)

	// ListAuthors lists people who have contributed (i.e., committed) code to
	// repo.
	ListAuthors(repo RepoRevSpec, opt *RepoListAuthorsOptions) ([]*AugmentedRepoAuthor, Response, error)

	// ListClients lists people who reference defs defined in repo.
	ListClients(repo RepoSpec, opt *RepoListClientsOptions) ([]*AugmentedRepoClient, Response, error)

	// ListDependents lists repositories that contain defs referenced by
	// repo.
	ListDependencies(repo RepoRevSpec, opt *RepoListDependenciesOptions) ([]*AugmentedRepoDependency, Response, error)

	// ListDependents lists repositories that reference defs defined in repo.
	ListDependents(repo RepoSpec, opt *RepoListDependentsOptions) ([]*AugmentedRepoDependent, Response, error)

	// ListByContributor lists repositories that person has contributed (i.e.,
	// committed) code to.
	ListByContributor(person PersonSpec, opt *RepoListByContributorOptions) ([]*AugmentedRepoContribution, Response, error)

	// ListByClient lists repositories that contain defs referenced by
	// person.
	ListByClient(person PersonSpec, opt *RepoListByClientOptions) ([]*AugmentedRepoUsageByClient, Response, error)

	// ListByRefdAuthor lists repositories that reference code authored by
	// person.
	ListByRefdAuthor(person PersonSpec, opt *RepoListByRefdAuthorOptions) ([]*AugmentedRepoUsageOfAuthor, Response, error)
}

// repositoriesService implements ReposService.
type repositoriesService struct {
	client *Client
}

var _ ReposService = &repositoriesService{}

// RepoSpec specifies a repository.
type RepoSpec struct {
	URI string
	RID int
}

// PathComponent returns the URL path component that specifies the person.
func (s RepoSpec) PathComponent() string {
	if s.RID > 0 {
		return "R$" + strconv.Itoa(s.RID)
	}
	if s.URI != "" {
		if strings.HasPrefix("sourcegraph.com/", s.URI) {
			return s.URI[len("sourcegraph.com/"):]
		} else {
			return s.URI
		}
	}
	panic("empty RepoSpec")
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
	if strings.HasPrefix(pathComponent, "R$") {
		rid, err := strconv.Atoi(pathComponent[2:])
		return RepoSpec{RID: rid}, err
	}

	var uri string
	if strings.HasPrefix(pathComponent, "sourcegraph/") {
		uri = "sourcegraph.com/" + pathComponent
	} else {
		uri = pathComponent
	}

	return RepoSpec{URI: uri}, nil
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
	RepoSpec        // repository URI or RID
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
type RepoGetOptions struct {
	Stats bool `url:",omitempty" json:",omitempty"` // whether to fetch and include stats in the returned repository
}

func (s *repositoriesService) Get(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error) {
	url, err := s.client.url(router.Repo, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repo_ *Repo
	resp, err := s.client.Do(req, &repo_)
	if err != nil {
		return nil, resp, err
	}

	return repo_, resp, nil
}

func (s *repositoriesService) GetStats(repoRev RepoRevSpec) (RepoStats, Response, error) {
	url, err := s.client.url(router.RepoStats, repoRev.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var stats RepoStats
	resp, err := s.client.Do(req, &stats)
	if err != nil {
		return nil, resp, err
	}

	return stats, resp, nil
}

func (s *repositoriesService) GetOrCreate(repo_ RepoSpec, opt *RepoGetOptions) (*Repo, Response, error) {
	url, err := s.client.url(router.ReposGetOrCreate, repo_.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repo__ *Repo
	resp, err := s.client.Do(req, &repo__)
	if err != nil {
		return nil, resp, err
	}

	return repo__, resp, nil
}

// RepoSettings describes a repository's configuration settings.
type RepoSettings struct {
	// BuildPushes is whether head commits on newly pushed branches
	// should be automatically built.
	BuildPushes *bool `db:"build_pushes" json:",omitempty"`

	SrcbotEnabled *bool `json:",omitempty"`

	// UseSSHPrivateKey is whether Sourcegraph should clone and update
	// the repository using an SSH key, and whether it should copy the
	// corresponding public key to the repository's origin host as an
	// authorized key. It is only necessary for private repositories
	// and for write operations on public repositories.
	UseSSHPrivateKey *bool `db:"use_ssh_private_key" json:",omitempty"`
}

func (s *repositoriesService) GetSettings(repo RepoSpec) (*RepoSettings, Response, error) {
	url, err := s.client.url(router.RepoSettings, repo.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var settings *RepoSettings
	resp, err := s.client.Do(req, &settings)
	if err != nil {
		return nil, resp, err
	}

	return settings, resp, nil
}

func (s *repositoriesService) UpdateSettings(repo RepoSpec, settings RepoSettings) (Response, error) {
	url, err := s.client.url(router.RepoSettingsUpdate, repo.RouteVars(), nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), settings)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *repositoriesService) RefreshProfile(repo RepoSpec) (Response, error) {
	url, err := s.client.url(router.RepoRefreshProfile, repo.RouteVars(), nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *repositoriesService) RefreshVCSData(repo RepoSpec) (Response, error) {
	url, err := s.client.url(router.RepoRefreshVCSData, repo.RouteVars(), nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *repositoriesService) ComputeStats(repo RepoRevSpec) (Response, error) {
	url, err := s.client.url(router.RepoComputeStats, repo.RouteVars(), nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// RepoGetBuildOptions sets options for the Repos.GetBuild call.
type RepoGetBuildOptions struct {
	// Exact is whether only a build whose commit ID exactly matches
	// the revspec should be returned. (For non-full-commit ID
	// revspecs, such as branches, tags, and partial commit IDs, this
	// means that the build's commit ID matches the resolved revspec's
	// commit ID.)
	//
	// If Exact is false, then builds for older commits that are
	// reachable from the revspec may also be returned. For example,
	// if there's a build for master~1 but no build for master, and
	// your revspec is master, using Exact=false will return the build
	// for master~1.
	//
	// Using Exact=true is faster as the commit and build history
	// never needs to be searched. If the exact build is not
	// found, or the exact build was found but it failed,
	// LastSuccessful and LastSuccessfulCommit for RepoBuildInfo
	// will be nil.
	Exact bool `url:",omitempty" json:",omitempty"`
}

// RepoBuildInfo holds a repository build (if one exists for the
// originally specified revspec) and additional information. It is returned by
// Repos.GetBuild.
type RepoBuildInfo struct {
	Exact *Build // the newest build, if any, that exactly matches the revspec (can be same as LastSuccessful)

	LastSuccessful *Build // the last successful build of a commit ID reachable from the revspec (can be same as Exact)

	CommitsBehind        int     // the number of commits between the revspec and the commit of the LastSuccessful build
	LastSuccessfulCommit *Commit // the commit of the LastSuccessful build
}

func (s *repositoriesService) GetBuild(repo RepoRevSpec, opt *RepoGetBuildOptions) (*RepoBuildInfo, Response, error) {
	url, err := s.client.url(router.RepoBuild, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var info *RepoBuildInfo
	resp, err := s.client.Do(req, &info)
	if err != nil {
		return nil, resp, err
	}

	return info, resp, nil
}

type NewRepoSpec struct {
	Type        string
	CloneURLStr string `json:"CloneURL"`
}

func (s *repositoriesService) Create(newRepoSpec NewRepoSpec) (*Repo, Response, error) {
	url, err := s.client.url(router.ReposCreate, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", url.String(), newRepoSpec)
	if err != nil {
		return nil, nil, err
	}

	var repo_ *Repo
	resp, err := s.client.Do(req, &repo_)
	if err != nil {
		return nil, resp, err
	}

	return repo_, resp, nil
}

func (s *repositoriesService) GetReadme(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error) {
	url, err := s.client.url(router.RepoReadme, repo.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var readme *vcsclient.TreeEntry
	resp, err := s.client.Do(req, &readme)
	if err != nil {
		return nil, resp, err
	}

	return readme, resp, nil
}

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

	Stats bool `url:",omitempty" json:",omitempty"` // whether to fetch and include stats in the returned repositories

	ListOptions
}

func (s *repositoriesService) List(opt *RepoListOptions) ([]*Repo, Response, error) {
	url, err := s.client.url(router.Repos, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repos []*Repo
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

type Commit struct {
	*vcs.Commit
}

type RepoListCommitsOptions struct {
	Head string `url:",omitempty" json:",omitempty"`
	ListOptions
}

func (s *repositoriesService) ListCommits(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error) {
	url, err := s.client.url(router.RepoCommits, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var commits []*Commit
	resp, err := s.client.Do(req, &commits)
	if err != nil {
		return nil, resp, err
	}

	return commits, resp, nil
}

type RepoGetCommitOptions struct {
}

func (s *repositoriesService) GetCommit(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error) {
	url, err := s.client.url(router.RepoCommit, rev.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var commit *Commit
	resp, err := s.client.Do(req, &commit)
	if err != nil {
		return nil, resp, err
	}

	return commit, resp, nil
}

type RepoListBranchesOptions struct {
	ListOptions
}

func (s *repositoriesService) ListBranches(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error) {
	url, err := s.client.url(router.RepoBranches, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var branches []*vcs.Branch
	resp, err := s.client.Do(req, &branches)
	if err != nil {
		return nil, resp, err
	}

	return branches, resp, nil
}

type RepoListTagsOptions struct {
	ListOptions
}

func (s *repositoriesService) ListTags(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error) {
	url, err := s.client.url(router.RepoTags, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var tags []*vcs.Tag
	resp, err := s.client.Do(req, &tags)
	if err != nil {
		return nil, resp, err
	}

	return tags, resp, nil
}

type Badge struct {
	Name              string
	Description       string
	ImageURL          string
	UncountedImageURL string
	Markdown          string
}

func (b *Badge) HTML() string {
	return fmt.Sprintf(`<img src="%s" alt="%s">`, template.HTMLEscapeString(b.ImageURL), template.HTMLEscapeString(b.Name))
}

func (s *repositoriesService) ListBadges(repo RepoSpec) ([]*Badge, Response, error) {
	url, err := s.client.url(router.RepoBadges, repo.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var badges []*Badge
	resp, err := s.client.Do(req, &badges)
	if err != nil {
		return nil, resp, err
	}

	return badges, resp, nil
}

type Counter struct {
	Name              string
	Description       string
	ImageURL          string
	UncountedImageURL string
	Markdown          string
}

func (c *Counter) HTML() string {
	return fmt.Sprintf(`<img src="%s" alt="%s">`, template.HTMLEscapeString(c.ImageURL), template.HTMLEscapeString(c.Name))
}

func (s *repositoriesService) ListCounters(repo RepoSpec) ([]*Counter, Response, error) {
	url, err := s.client.url(router.RepoCounters, repo.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var counters []*Counter
	resp, err := s.client.Do(req, &counters)
	if err != nil {
		return nil, resp, err
	}

	return counters, resp, nil
}

type RepoAuthor struct {
	UID   nnz.Int
	Email nnz.String
	AuthorStats
}

// AugmentedRepoAuthor is a RepoAuthor with the full User and
// graph.Def structs embedded.
type AugmentedRepoAuthor struct {
	User *User
	*RepoAuthor
}

type RepoListAuthorsOptions struct {
	ListOptions
}

func (s *repositoriesService) ListAuthors(repo RepoRevSpec, opt *RepoListAuthorsOptions) ([]*AugmentedRepoAuthor, Response, error) {
	url, err := s.client.url(router.RepoAuthors, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var authors []*AugmentedRepoAuthor
	resp, err := s.client.Do(req, &authors)
	if err != nil {
		return nil, resp, err
	}

	return authors, resp, nil
}

type RepoClient struct {
	UID   nnz.Int
	Email nnz.String
	ClientStats
}

type ClientStats struct {
	AuthorshipInfo

	// DefRepo is the URI of the repository that defines defs that
	// this client referred to.
	DefRepo string `db:"def_repo"`

	// DefUnitType and DefUnit are the unit in DefRepo that defines
	// defs that this client referred to. If DefUnitType == "" and
	// DefUnit == "", then this ClientStats is an aggregate of this client's
	// refs to all units in DefRepo.
	DefUnitType nnz.String `db:"def_unit_type"`
	DefUnit     nnz.String `db:"def_unit"`

	// RefCount is the number of references this client made in this repository
	// to DefRepo.
	RefCount int `db:"ref_count"`
}

// AugmentedRepoClient is a RepoClient with the full User and
// graph.Def structs embedded.
type AugmentedRepoClient struct {
	User *User
	*RepoClient
}

type RepoListClientsOptions struct {
	ListOptions
}

func (s *repositoriesService) ListClients(repo RepoSpec, opt *RepoListClientsOptions) ([]*AugmentedRepoClient, Response, error) {
	url, err := s.client.url(router.RepoClients, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var clients []*AugmentedRepoClient
	resp, err := s.client.Do(req, &clients)
	if err != nil {
		return nil, resp, err
	}

	return clients, resp, nil
}

type RepoDependency struct {
	ToRepo string `db:"to_repo"`
}

type AugmentedRepoDependency struct {
	Repo *Repo
	*RepoDependency
}

type RepoListDependenciesOptions struct {
	ListOptions
}

func (s *repositoriesService) ListDependencies(repo RepoRevSpec, opt *RepoListDependenciesOptions) ([]*AugmentedRepoDependency, Response, error) {
	url, err := s.client.url(router.RepoDependencies, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var dependencies []*AugmentedRepoDependency
	resp, err := s.client.Do(req, &dependencies)
	if err != nil {
		return nil, resp, err
	}

	return dependencies, resp, nil
}

type RepoDependent struct {
	FromRepo string `db:"from_repo"`
}

type AugmentedRepoDependent struct {
	Repo *Repo
	*RepoDependent
}

type RepoListDependentsOptions struct{ ListOptions }

func (s *repositoriesService) ListDependents(repo RepoSpec, opt *RepoListDependentsOptions) ([]*AugmentedRepoDependent, Response, error) {
	url, err := s.client.url(router.RepoDependents, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var dependents []*AugmentedRepoDependent
	resp, err := s.client.Do(req, &dependents)
	if err != nil {
		return nil, resp, err
	}

	return dependents, resp, nil
}

type AuthorStats struct {
	AuthorshipInfo

	// DefCount is the number of defs that this author contributed (where
	// "contributed" means "committed any hunk of code to source code files").
	DefCount int `db:"def_count"`

	DefsProportion float64 `db:"defs_proportion"`

	// ExportedDefCount is the number of exported defs that this author
	// contributed (where "contributed to" means "committed any hunk of code to
	// source code files").
	ExportedDefCount int `db:"exported_def_count"`

	ExportedDefsProportion float64 `db:"exported_defs_proportion"`

	// TODO(sqs): add "most recently contributed exported def"
}

type RepoContribution struct {
	RepoURI string `db:"repo"`
	AuthorStats
}

type AugmentedRepoContribution struct {
	Repo *Repo
	*RepoContribution
}

type RepoListByContributorOptions struct {
	NoFork bool
	ListOptions
}

func (s *repositoriesService) ListByContributor(person PersonSpec, opt *RepoListByContributorOptions) ([]*AugmentedRepoContribution, Response, error) {
	url, err := s.client.url(router.PersonRepoContributions, person.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repos []*AugmentedRepoContribution
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

// RepoUsageByClient describes a repository whose code is referenced by a
// specific person.
type RepoUsageByClient struct {
	// DefRepo is the repository that defines the code that was referenced.
	// It's called DefRepo because "Repo" usually refers to the repository
	// whose analysis created this linkage (i.e., the repository that contains
	// the reference).
	DefRepo string `db:"def_repo"`

	RefCount int `db:"ref_count"`

	AuthorshipInfo
}

// AugmentedRepoUsageByClient is a RepoUsageByClient with the full Repo
// struct embedded.
type AugmentedRepoUsageByClient struct {
	DefRepo            *Repo
	*RepoUsageByClient `json:"RepoUsageByClient"`
}

type RepoListByClientOptions struct {
	ListOptions
}

func (s *repositoriesService) ListByClient(person PersonSpec, opt *RepoListByClientOptions) ([]*AugmentedRepoUsageByClient, Response, error) {
	url, err := s.client.url(router.PersonRepoDependencies, person.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repos []*AugmentedRepoUsageByClient
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

// RepoUsageOfAuthor describes a repository referencing code committed by a
// specific person.
type RepoUsageOfAuthor struct {
	Repo string

	RefCount int `db:"ref_count"`
}

// AugmentedRepoUsageOfAuthor is a RepoUsageOfAuthor with the full
// Repo struct embedded.
type AugmentedRepoUsageOfAuthor struct {
	Repo               *Repo
	*RepoUsageOfAuthor `json:"RepoUsageOfAuthor"`
}

type RepoListByRefdAuthorOptions struct {
	ListOptions
}

func (s *repositoriesService) ListByRefdAuthor(person PersonSpec, opt *RepoListByRefdAuthorOptions) ([]*AugmentedRepoUsageOfAuthor, Response, error) {
	url, err := s.client.url(router.PersonRepoDependents, person.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var repos []*AugmentedRepoUsageOfAuthor
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

type MockReposService struct {
	Get_               func(spec RepoSpec, opt *RepoGetOptions) (*Repo, Response, error)
	GetStats_          func(repo RepoRevSpec) (RepoStats, Response, error)
	GetOrCreate_       func(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error)
	GetSettings_       func(repo RepoSpec) (*RepoSettings, Response, error)
	UpdateSettings_    func(repo RepoSpec, settings RepoSettings) (Response, error)
	RefreshProfile_    func(repo RepoSpec) (Response, error)
	RefreshVCSData_    func(repo RepoSpec) (Response, error)
	ComputeStats_      func(repo RepoRevSpec) (Response, error)
	GetBuild_          func(repo RepoRevSpec, opt *RepoGetBuildOptions) (*RepoBuildInfo, Response, error)
	Create_            func(newRepoSpec NewRepoSpec) (*Repo, Response, error)
	GetReadme_         func(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error)
	List_              func(opt *RepoListOptions) ([]*Repo, Response, error)
	ListCommits_       func(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error)
	GetCommit_         func(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error)
	ListBranches_      func(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error)
	ListTags_          func(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error)
	ListBadges_        func(repo RepoSpec) ([]*Badge, Response, error)
	ListCounters_      func(repo RepoSpec) ([]*Counter, Response, error)
	ListAuthors_       func(repo RepoRevSpec, opt *RepoListAuthorsOptions) ([]*AugmentedRepoAuthor, Response, error)
	ListClients_       func(repo RepoSpec, opt *RepoListClientsOptions) ([]*AugmentedRepoClient, Response, error)
	ListDependencies_  func(repo RepoRevSpec, opt *RepoListDependenciesOptions) ([]*AugmentedRepoDependency, Response, error)
	ListDependents_    func(repo RepoSpec, opt *RepoListDependentsOptions) ([]*AugmentedRepoDependent, Response, error)
	ListByContributor_ func(person PersonSpec, opt *RepoListByContributorOptions) ([]*AugmentedRepoContribution, Response, error)
	ListByClient_      func(person PersonSpec, opt *RepoListByClientOptions) ([]*AugmentedRepoUsageByClient, Response, error)
	ListByRefdAuthor_  func(person PersonSpec, opt *RepoListByRefdAuthorOptions) ([]*AugmentedRepoUsageOfAuthor, Response, error)
}

var _ ReposService = MockReposService{}

func (s MockReposService) Get(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error) {
	return s.Get_(repo, opt)
}

func (s MockReposService) GetStats(repo RepoRevSpec) (RepoStats, Response, error) {
	return s.GetStats_(repo)
}

func (s MockReposService) GetOrCreate(repo RepoSpec, opt *RepoGetOptions) (*Repo, Response, error) {
	return s.GetOrCreate_(repo, opt)
}

func (s MockReposService) GetSettings(repo RepoSpec) (*RepoSettings, Response, error) {
	return s.GetSettings_(repo)
}

func (s MockReposService) UpdateSettings(repo RepoSpec, settings RepoSettings) (Response, error) {
	return s.UpdateSettings_(repo, settings)
}

func (s MockReposService) RefreshProfile(repo RepoSpec) (Response, error) {
	return s.RefreshProfile_(repo)
}

func (s MockReposService) RefreshVCSData(repo RepoSpec) (Response, error) {
	return s.RefreshVCSData_(repo)
}

func (s MockReposService) ComputeStats(repo RepoRevSpec) (Response, error) {
	return s.ComputeStats_(repo)
}

func (s MockReposService) GetBuild(repo RepoRevSpec, opt *RepoGetBuildOptions) (*RepoBuildInfo, Response, error) {
	return s.GetBuild_(repo, opt)
}

func (s MockReposService) Create(newRepoSpec NewRepoSpec) (*Repo, Response, error) {
	return s.Create_(newRepoSpec)
}

func (s MockReposService) GetReadme(repo RepoRevSpec) (*vcsclient.TreeEntry, Response, error) {
	return s.GetReadme_(repo)
}

func (s MockReposService) List(opt *RepoListOptions) ([]*Repo, Response, error) {
	return s.List_(opt)
}

func (s MockReposService) ListBadges(repo RepoSpec) ([]*Badge, Response, error) {
	return s.ListBadges_(repo)
}

func (s MockReposService) ListCounters(repo RepoSpec) ([]*Counter, Response, error) {
	return s.ListCounters_(repo)
}

func (s MockReposService) ListAuthors(repo RepoRevSpec, opt *RepoListAuthorsOptions) ([]*AugmentedRepoAuthor, Response, error) {
	return s.ListAuthors_(repo, opt)
}

func (s MockReposService) ListClients(repo RepoSpec, opt *RepoListClientsOptions) ([]*AugmentedRepoClient, Response, error) {
	return s.ListClients_(repo, opt)
}

func (s MockReposService) ListDependencies(repo RepoRevSpec, opt *RepoListDependenciesOptions) ([]*AugmentedRepoDependency, Response, error) {
	return s.ListDependencies_(repo, opt)
}

func (s MockReposService) ListDependents(repo RepoSpec, opt *RepoListDependentsOptions) ([]*AugmentedRepoDependent, Response, error) {
	return s.ListDependents_(repo, opt)
}

func (s MockReposService) ListByContributor(person PersonSpec, opt *RepoListByContributorOptions) ([]*AugmentedRepoContribution, Response, error) {
	return s.ListByContributor_(person, opt)
}

func (s MockReposService) ListByClient(person PersonSpec, opt *RepoListByClientOptions) ([]*AugmentedRepoUsageByClient, Response, error) {
	return s.ListByClient_(person, opt)
}

func (s MockReposService) ListByRefdAuthor(person PersonSpec, opt *RepoListByRefdAuthorOptions) ([]*AugmentedRepoUsageOfAuthor, Response, error) {
	return s.ListByRefdAuthor_(person, opt)
}

func (s MockReposService) ListCommits(repo RepoSpec, opt *RepoListCommitsOptions) ([]*Commit, Response, error) {
	return s.ListCommits_(repo, opt)
}

func (s MockReposService) GetCommit(rev RepoRevSpec, opt *RepoGetCommitOptions) (*Commit, Response, error) {
	return s.GetCommit_(rev, opt)
}

func (s MockReposService) ListBranches(repo RepoSpec, opt *RepoListBranchesOptions) ([]*vcs.Branch, Response, error) {
	return s.ListBranches_(repo, opt)
}

func (s MockReposService) ListTags(repo RepoSpec, opt *RepoListTagsOptions) ([]*vcs.Tag, Response, error) {
	return s.ListTags_(repo, opt)
}
