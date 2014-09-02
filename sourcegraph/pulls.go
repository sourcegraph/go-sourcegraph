package sourcegraph

import (
	"github.com/sqs/go-github/github"

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
}

// pullRequestsService implements PullRequestsService.
type pullRequestsService struct {
	client *Client
}

var _ PullRequestsService = &pullRequestsService{}

// PullRequestSpec specifies a pull request.
type PullRequestSpec struct {
	Repo RepoSpec

	Number int // Sequence number of the pull request
}

func (s PullRequestSpec) RouteVars() map[string]string {
	return map[string]string{"RepoSpec": s.Repo.URI, "PullNumber": strconv.Itoa(s.Number)}
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

type MockPullRequestsService struct {
	Get_              func(pull PullRequestSpec, opt *PullRequestGetOptions) (*PullRequest, Response, error)
	ListByRepository_ func(repo RepoSpec, opt *PullRequestListOptions) ([]*PullRequest, Response, error)
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
