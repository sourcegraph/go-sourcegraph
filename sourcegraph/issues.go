package sourcegraph

import (
	"fmt"

	"github.com/sqs/go-github/github"

	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// IssuesService communicates with the issue-related endpoints in the
// Sourcegraph API.
type IssuesService interface {
	// Get fetches a issue.
	Get(issue IssueSpec, opt *IssueGetOptions) (*Issue, Response, error)

	// List issues for a repository.
	ListByRepository(repo RepoSpec, opt *IssueListOptions) ([]*Issue, Response, error)

	// ListComments lists comments on a issue.
	ListComments(issue IssueSpec, opt *IssueListCommentsOptions) ([]*IssueComment, Response, error)
}

// issuesService implements IssuesService.
type issuesService struct {
	client *Client
}

var _ IssuesService = &issuesService{}

// IssueSpec specifies a issue.
type IssueSpec struct {
	Repo RepoSpec

	Number int // Sequence number of the issue
}

func (s IssueSpec) RouteVars() map[string]string {
	return map[string]string{"RepoSpec": s.Repo.URI, "Issue": strconv.Itoa(s.Number)}
}

func UnmarshalIssueSpec(routeVars map[string]string) (IssueSpec, error) {
	issueNumber, err := strconv.Atoi(routeVars["Issue"])
	if err != nil {
		return IssueSpec{}, err
	}
	repoURI := routeVars["RepoSpec"]
	if repoURI == "" {
		return IssueSpec{}, fmt.Errorf("RepoSpec was empty")
	}
	return IssueSpec{
		Repo:   RepoSpec{URI: repoURI},
		Number: issueNumber,
	}, nil
}

// Issue is a issue returned by the Sourcegraph API.
type Issue struct {
	github.Issue
}

// Spec returns the IssueSpec that specifies r.
func (r *Issue) Spec() IssueSpec {
	// Extract the URI from the HTMLURL field.
	uri := strings.Join(strings.Split(strings.TrimPrefix(*r.HTMLURL, "https://"), "/")[0:3], "/")
	return IssueSpec{
		Repo:   RepoSpec{URI: uri},
		Number: *r.Number,
	}
}

type IssueGetOptions struct{}

func (s *issuesService) Get(issue IssueSpec, opt *IssueGetOptions) (*Issue, Response, error) {
	url, err := s.client.url(router.RepoIssue, issue.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var issue_ *Issue
	resp, err := s.client.Do(req, &issue_)
	if err != nil {
		return nil, resp, err
	}

	return issue_, resp, nil
}

type IssueListOptions struct {
	State string `url:",omitempty"` // "open", "closed", or "all"
	ListOptions
}

func (s *issuesService) ListByRepository(repo RepoSpec, opt *IssueListOptions) ([]*Issue, Response, error) {
	url, err := s.client.url(router.RepoIssues, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var issues []*Issue
	resp, err := s.client.Do(req, &issues)
	if err != nil {
		return nil, resp, err
	}

	return issues, resp, nil
}

type IssueListCommentsOptions struct {
	ListOptions
}

type IssueComment struct {
	github.IssueComment
}

func (s *issuesService) ListComments(issue IssueSpec, opt *IssueListCommentsOptions) ([]*IssueComment, Response, error) {
	url, err := s.client.url(router.RepoIssueComments, issue.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var comments []*IssueComment
	resp, err := s.client.Do(req, &comments)
	if err != nil {
		return nil, resp, err
	}

	return comments, resp, nil
}

type MockIssuesService struct {
	Get_              func(issue IssueSpec, opt *IssueGetOptions) (*Issue, Response, error)
	ListByRepository_ func(repo RepoSpec, opt *IssueListOptions) ([]*Issue, Response, error)
	ListComments_     func(issue IssueSpec, opt *IssueListCommentsOptions) ([]*IssueComment, Response, error)
}

var _ IssuesService = MockIssuesService{}

func (s MockIssuesService) Get(issue IssueSpec, opt *IssueGetOptions) (*Issue, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(issue, opt)
}

func (s MockIssuesService) ListByRepository(repo RepoSpec, opt *IssueListOptions) ([]*Issue, Response, error) {
	if s.ListByRepository_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListByRepository_(repo, opt)
}

func (s MockIssuesService) ListComments(issue IssueSpec, opt *IssueListCommentsOptions) ([]*IssueComment, Response, error) {
	if s.ListComments_ == nil {
		return nil, nil, nil
	}
	return s.ListComments_(issue, opt)
}
