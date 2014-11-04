package sourcegraph

import (
	"github.com/sourcegraph/go-github/github"

	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// PullRequestsService communicates with the pull request-related endpoints in the
// Sourcegraph API.
type PullRequestsService interface {
	// Get fetches a pull request.
	Get(pull PullRequestSpec, opt *PullRequestGetOptions) (*PullRequest, Response, error)

	// List pull requests for a repository.
	ListByRepository(repo RepoSpec, opt *PullRequestListOptions) ([]*PullRequest, Response, error)

	// ListComments lists comments on a pull request.
	ListComments(pull PullRequestSpec, opt *PullRequestListCommentsOptions) ([]*PullRequestComment, Response, error)
}

// pullRequestsService implements PullRequestsService.
type pullRequestsService struct {
	client *Client
}

var _ PullRequestsService = &pullRequestsService{}

// PullRequestSpec specifies a pull request.
type PullRequestSpec struct {
	Repo RepoSpec // the base repository of the pull request

	Number int // Sequence number of the pull request
}

// RouteVars returns the route variables for generating pull request
// URLs.
func (s PullRequestSpec) RouteVars() map[string]string {
	return map[string]string{"RepoSpec": s.Repo.URI, "Pull": strconv.Itoa(s.Number)}
}

// IssueSpec returns a specifier for the issue associated with this
// pull request (same repo, same number).
func (s PullRequestSpec) IssueSpec() IssueSpec {
	return IssueSpec{Repo: s.Repo, Number: s.Number}
}

// UnmarshalPullRequestSpec parses route variables (a map returned by
// (PullRequestSpec).RouteVars()) to construct a PullRequestSpec.
func UnmarshalPullRequestSpec(v map[string]string) (PullRequestSpec, error) {
	ps := PullRequestSpec{}
	var err error
	ps.Repo, err = UnmarshalRepoSpec(v)
	if err != nil {
		return ps, err
	}

	ps.Number, err = strconv.Atoi(v["Pull"])
	return ps, err
}

// PullRequest is a pull request returned by the Sourcegraph API.
type PullRequest struct {
	github.PullRequest
}

// Spec returns the PullRequestSpec that specifies r.
func (r *PullRequest) Spec() PullRequestSpec {
	// Extract the URI from the HTMLURL field.
	uri := strings.Join(strings.Split(strings.TrimPrefix(*r.HTMLURL, "https://"), "/")[0:3], "/")
	return PullRequestSpec{
		Repo:   RepoSpec{URI: uri},
		Number: *r.Number,
	}
}

type PullRequestGetOptions struct{}

func (s *pullRequestsService) Get(pull PullRequestSpec, opt *PullRequestGetOptions) (*PullRequest, Response, error) {
	url, err := s.client.url(router.RepoPullRequest, pull.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var pull_ *PullRequest
	resp, err := s.client.Do(req, &pull_)
	if err != nil {
		return nil, resp, err
	}

	return pull_, resp, nil
}

type PullRequestListOptions struct {
	State string `url:",omitempty"` // "open", "closed", or "all"
	ListOptions
}

func (s *pullRequestsService) ListByRepository(repo RepoSpec, opt *PullRequestListOptions) ([]*PullRequest, Response, error) {
	url, err := s.client.url(router.RepoPullRequests, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var pulls []*PullRequest
	resp, err := s.client.Do(req, &pulls)
	if err != nil {
		return nil, resp, err
	}

	return pulls, resp, nil
}

type PullRequestListCommentsOptions struct {
	ListOptions
}

type PullRequestComment struct {
	github.PullRequestComment
}

func (s *pullRequestsService) ListComments(pull PullRequestSpec, opt *PullRequestListCommentsOptions) ([]*PullRequestComment, Response, error) {
	url, err := s.client.url(router.RepoPullRequestComments, pull.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var comments []*PullRequestComment
	resp, err := s.client.Do(req, &comments)
	if err != nil {
		return nil, resp, err
	}

	return comments, resp, nil
}

type MockPullRequestsService struct {
	Get_              func(pull PullRequestSpec, opt *PullRequestGetOptions) (*PullRequest, Response, error)
	ListByRepository_ func(repo RepoSpec, opt *PullRequestListOptions) ([]*PullRequest, Response, error)
	ListComments_     func(pull PullRequestSpec, opt *PullRequestListCommentsOptions) ([]*PullRequestComment, Response, error)
}

var _ PullRequestsService = MockPullRequestsService{}

func (s MockPullRequestsService) Get(pull PullRequestSpec, opt *PullRequestGetOptions) (*PullRequest, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(pull, opt)
}

func (s MockPullRequestsService) ListByRepository(repo RepoSpec, opt *PullRequestListOptions) ([]*PullRequest, Response, error) {
	if s.ListByRepository_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListByRepository_(repo, opt)
}

func (s MockPullRequestsService) ListComments(pull PullRequestSpec, opt *PullRequestListCommentsOptions) ([]*PullRequestComment, Response, error) {
	if s.ListComments_ == nil {
		return nil, nil, nil
	}
	return s.ListComments_(pull, opt)
}
