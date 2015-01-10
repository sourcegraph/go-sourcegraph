package sourcegraph

import (
	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

type CombinedStatus struct {
	github.CombinedStatus
}

func (s *repositoriesService) GetCombinedStatus(spec RepoRevSpec, opt *ListOptions) (*CombinedStatus, Response, error) {
	url, err := s.client.URL(router.RepoCombinedStatus, spec.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var status CombinedStatus
	resp, err := s.client.Do(req, &status)
	if err != nil {
		return nil, resp, err
	}

	return &status, resp, nil
}
