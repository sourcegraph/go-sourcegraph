package sourcegraph

import (
	"fmt"

	"github.com/sourcegraph/vcsstore/vcsclient"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// RepositoryTreeService communicates with the Sourcegraph API endpoints that
// fetch file and directory entries in repositories.
type RepositoryTreeService interface {
	Get(entry TreeEntrySpec, opt *RepositoryTreeGetOptions) (*TreeEntry, Response, error)
}

type repositoryTreeService struct {
	client *Client
}

var _ RepositoryTreeService = &repositoryTreeService{}

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

	// The line in the file that the formatted section starts at
	StartLine int

	// The line that the formatted section ends at
	EndLine int
}

// RepositoryTreeGetOptions specifies options for (RepositoryTreeService).Get.
type RepositoryTreeGetOptions struct {
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

	CodeFormatOptions
}

type CodeFormatOptions struct {
	// StartLine and EndLine, if EndLine is nonzero, specify the line range of
	// the file to fetch.
	StartLine int `url:",omitempty"`
	EndLine   int `url:",omitempty"`

	// EntireFile is whether the entire file contents should be annotated. If
	// true, Start and End are ignored.
	EntireFile bool `url:",omitempty"`

	// LineNumberedTableRows is whether to wrap each line in a <tr> element.
	LineNumberedTableRows bool `url:",omitempty"`

	// StartByte and EndByte, if EndByte is nonzero, specify the byte range of
	// the file to fetch.
	StartByte, EndByte int

	// ExpandContextLines is how many lines of output context to include (if
	// StartByte and EndByte are specified). For
	// example, specifying 2 will expand the annotation range to include 2 full
	// lines before the beginning and 2 full lines after the end.
	ExpandContextLines int `url:",omitempty"`

	// FullLines is whether an annotation range that includes partial lines
	// should be extended to the nearest line boundaries on both sides. It is
	// only valid if StartByte and EndByte are specified.
	FullLines bool `url:",omitempty"`
}

func (s *repositoryTreeService) Get(entry TreeEntrySpec, opt *RepositoryTreeGetOptions) (*TreeEntry, Response, error) {
	url, err := s.client.url(router.RepositoryTreeEntry, entry.RouteVars(), opt)
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

type MockRepositoryTreeService struct {
	Get_ func(entry TreeEntrySpec, opt *RepositoryTreeGetOptions) (*TreeEntry, Response, error)
}

var _ RepositoryTreeService = MockRepositoryTreeService{}

func (s MockRepositoryTreeService) Get(entry TreeEntrySpec, opt *RepositoryTreeGetOptions) (*TreeEntry, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(entry, opt)
}
