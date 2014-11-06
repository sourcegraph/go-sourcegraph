package sourcegraph

import (
	"io"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

// BuildDataService communicates with the build data-related endpoints in the
// Sourcegraph API.
type BuildDataService interface {
	// List lists build data files and subdirectories.
	List(repo RepoRevSpec, opt *BuildDataListOptions) ([]*buildstore.BuildDataFileInfo, Response, error)

	// Get gets a build data file.
	Get(file BuildDataFileSpec) (io.ReadCloser, Response, error)

	// Upload uploads a build data file.
	Upload(spec BuildDataFileSpec, body io.ReadCloser) (Response, error)
}

type buildDataService struct {
	client *Client
}

var _ BuildDataService = &buildDataService{}

// BuildDataFileSpec specifies a new or existing build data file in a
// repository.
type BuildDataFileSpec struct {
	RepoRev RepoRevSpec
	Path    string
}

// RouteVars returns route variables used to construct URLs to a build
// data file.
func (s *BuildDataFileSpec) RouteVars() map[string]string {
	m := s.RepoRev.RouteVars()
	m["Path"] = s.Path
	return m
}

const (
	// MIME types to use in Accept request header for the server to
	// know (without statting the path) what kind of resource (file or
	// directory) the client wants to fetch.
	BuildDataFileContentType = "application/vnd.sourcegraph.build-data-file"
	BuildDataDirContentType  = "application/vnd.sourcegraph.build-data-dir"
)

// BuildDataListOptions specifies options for listing build data
// files.
type BuildDataListOptions struct {
	ListOptions
}

func (s *buildDataService) List(repo RepoRevSpec, opt *BuildDataListOptions) ([]*buildstore.BuildDataFileInfo, Response, error) {
	v := repo.RouteVars()
	v["Path"] = "."
	url, err := s.client.url(router.RepositoryBuildDataEntry, v, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("accept", BuildDataDirContentType)

	var fileInfo []*buildstore.BuildDataFileInfo
	resp, err := s.client.Do(req, &fileInfo)
	if err != nil {
		return nil, resp, err
	}

	return fileInfo, resp, nil
}

func (s *buildDataService) Get(file BuildDataFileSpec) (io.ReadCloser, Response, error) {
	url, err := s.client.url(router.RepositoryBuildDataEntry, file.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("accept", BuildDataFileContentType)

	resp, err := s.client.Do(req, preserveBody)
	if err != nil {
		return nil, resp, err
	}

	return resp.Body, resp, nil
}

func (s *buildDataService) Upload(file BuildDataFileSpec, body io.ReadCloser) (Response, error) {
	url, err := s.client.url(router.RepositoryBuildDataEntry, file.RouteVars(), nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Body = body

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

type MockBuildDataService struct {
	List_   func(repo RepoRevSpec, opt *BuildDataListOptions) ([]*buildstore.BuildDataFileInfo, Response, error)
	Get_    func(file BuildDataFileSpec) (io.ReadCloser, Response, error)
	Upload_ func(spec BuildDataFileSpec, body io.ReadCloser) (Response, error)
}

var _ BuildDataService = MockBuildDataService{}

func (s MockBuildDataService) List(repo RepoRevSpec, opt *BuildDataListOptions) ([]*buildstore.BuildDataFileInfo, Response, error) {
	if s.List_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.List_(repo, opt)
}

func (s MockBuildDataService) Get(file BuildDataFileSpec) (io.ReadCloser, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(file)
}

func (s MockBuildDataService) Upload(spec BuildDataFileSpec, body io.ReadCloser) (Response, error) {
	if s.Upload_ == nil {
		return nil, nil
	}
	return s.Upload_(spec, body)
}
