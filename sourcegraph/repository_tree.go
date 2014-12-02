package sourcegraph

import (
	"fmt"

	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// RepoTreeService communicates with the Sourcegraph API endpoints that
// fetch file and directory entries in repositories.
type RepoTreeService interface {
	Get(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error)
}

type repoTreeService struct {
	client *Client
}

var _ RepoTreeService = &repoTreeService{}

type TreeEntrySpec struct {
	RepoRev RepoRevSpec
	Path    string
}

func (s *TreeEntrySpec) RouteVars() map[string]string {
	m := s.RepoRev.RouteVars()
	m["Path"] = s.Path
	return m
}

func (s TreeEntrySpec) String() string {
	return fmt.Sprintf("%v: %s (rev %q)", s.RepoRev, s.Path, s.RepoRev.Rev)
}

// TreeEntry is a file or directory in a repository, with additional feedback
// from the formatting operation (if Formatted is true in the options).
type TreeEntry struct {
	*vcsclient.TreeEntry

	*vcsclient.FileRange // only set for files

	ContentsString string

	// FormatResult is only set if this TreeEntry is a file.
	FormatResult *FormatResult `json:",omitempty"`

	// EntryDefinitions is a list of defined defs for each entry in this
	// directory. It is only populated if DirEntryDefinitions is true.
	EntryDefinitions map[string]interface{}
}

// FormatResult contains information about and warnings from the formatting
// operation (if Formatted is true in the options).
type FormatResult struct {
	// TooManyRefs indicates that the file being formatted exceeded the maximum
	// number of refs that are linked. Only the first NumRefs refs are linked.
	TooManyRefs bool `json:",omitempty"`

	// NumRefs is the number of refs that were linked in this file. If the total
	// number of refs in the file exceeds the (server-defined) limit, NumRefs is
	// capped at the limit.
	NumRefs int

	// LineStartByteOffsets is the byte offset of each line's first
	// byte.
	LineStartByteOffsets []int
}

// RepoTreeGetOptions specifies options for (RepoTreeService).Get.
type RepoTreeGetOptions struct {
	// Formatted is whether the specified entry, if it's a file, should have its
	// contents code-formatted.
	Formatted bool

	// DirEntryDefinitions is whether the specified entry, if it's a directory,
	// should include a list of defined defs for each of its entries (in
	// EntryDefinitions). For example, if the specified entry has a file "a" and
	// a dir "b/", the result would include a list of defs defined in "a" and
	// in any file underneath "b/". Not all defs defined in the entries are
	// returned; only the top few are.
	DirEntryDefinitions bool `url:",omitempty"`

	ContentsAsString bool `url:",omitempty"`

	vcsclient.GetFileOptions
}

func (s *repoTreeService) Get(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error) {
	url, err := s.client.url(router.RepoTreeEntry, entry.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var entry_ *TreeEntry
	resp, err := s.client.Do(req, &entry_)
	if err != nil {
		return nil, resp, err
	}

	return entry_, resp, nil
}

type MockRepoTreeService struct {
	Get_ func(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error)
}

var _ RepoTreeService = MockRepoTreeService{}

func (s MockRepoTreeService) Get(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(entry, opt)
}
