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

	ContentsString string `json:",omitempty"`

	// SourceCode contains the tokenized source code. This structure is only filled in
	// when "TokenizedSource" is set to "true" in RepoTreeGetOptions.
	SourceCode *SourceCode `json:",omitempty"`

	// FormatResult is only set if this TreeEntry is a file.
	FormatResult *FormatResult `json:",omitempty"`
}

// SourceCode contains a snippet of code with linked and classed tokens,
// along with other information about the contents.
type SourceCode struct {
	// Lines contains all the lines of the contained code snippet.
	Lines []*SourceCodeLine `json:"lines,omitempty"`

	NumRefs              int   `json:"refs"`
	TooManyRefs          bool  `json:"max"`
	LineStartByteOffsets []int `json:"lineOffsets"`
}

// SourceCodeLine contains all tokens on this line along with other information
// such as byte offsets in original source.
type SourceCodeLine struct {
	// StartByte and EndByte are the start and end offsets in bytes, in the original file.
	StartByte int `json:"s"`
	EndByte   int `json:"e"`

	// Tokens contains any tokens that may be on this line, including whitespace. Whitespace
	// is stored as an HTML encoded "string" and token information is stored as
	// "SourceCodeToken". New lines ('\n') are not present.
	Tokens []interface{} `json:"t,omitempty"`
}

// SourceCodeToken contains information about a code token.
type SourceCodeToken struct {
	// Start and end byte offsets in original file.
	StartByte int `json:"-"`
	EndByte   int `json:"-"`

	// URL specifies that the token is a reference or a definition,  based on the
	// IsDef property.
	URL string `json:"u,omitempty"`

	// IsDef specifies whether the token is a definition.
	IsDef bool `json:"d,omitempty"`

	// Class specifies the token type as per
	// [google-code-prettify](https://code.google.com/p/google-code-prettify/).
	Class string `json:"s"`

	// Label is non-whitespace HTML encoded source code.
	Label string `json:"h"`
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
	// Contents code-formatted using HTML.
	Formatted bool

	// TokenizedSource requests that the source code be returned as a data structure,
	// rather than an (annotated) string. This is useful when full control of rendering
	// and traversal of source code is desired on the client.
	TokenizedSource bool `url:",omitempty"`

	ContentsAsString bool `url:",omitempty"`

	vcsclient.GetFileOptions
}

func (s *repoTreeService) Get(entry TreeEntrySpec, opt *RepoTreeGetOptions) (*TreeEntry, Response, error) {
	url, err := s.client.URL(router.RepoTreeEntry, entry.RouteVars(), opt)
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
	return s.Get_(entry, opt)
}
