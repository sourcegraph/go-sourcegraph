package sourcegraph

import (
	"fmt"
	"html/template"
	"time"

	"github.com/sourcegraph/go-nnz/nnz"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/person"
	"sourcegraph.com/sourcegraph/srclib/repo"
)

// DefsService communicates with the def- and graph-related endpoints in
// the Sourcegraph API.
type DefsService interface {
	// Get fetches a def.
	Get(def DefSpec, opt *DefGetOptions) (*Def, Response, error)

	// List defs.
	List(opt *DefListOptions) ([]*Def, Response, error)

	// ListExamples lists examples for def.
	ListExamples(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error)

	// ListExamples lists people who committed parts of def's definition.
	ListAuthors(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error)

	// ListClients lists people who use def in their code.
	ListClients(def DefSpec, opt *DefListClientsOptions) ([]*AugmentedDefClient, Response, error)

	// ListDependents lists repositories that use def in their code.
	ListDependents(def DefSpec, opt *DefListDependentsOptions) ([]*AugmentedDefDependent, Response, error)

	// ListVersions lists all available versions of a definition in
	// the various repository commits in which it has appeared.
	//
	// TODO(sqs): how to deal with renames, etc.?
	ListVersions(def DefSpec, opt *DefListVersionsOptions) ([]*Def, Response, error)
}

// DefSpec specifies a def.
type DefSpec struct {
	Repo     string
	CommitID string
	UnitType string
	Unit     string
	Path     string
}

func (s *DefSpec) RouteVars() map[string]string {
	m := map[string]string{"RepoSpec": s.Repo, "UnitType": s.UnitType, "Unit": s.Unit, "Path": s.Path}
	if s.CommitID != "" {
		m["Rev"] = s.CommitID
	}
	return m
}

// DefKey returns the def key specified by s, using the Repo, UnitType,
// Unit, and Path fields of s.
func (s *DefSpec) DefKey() graph.DefKey {
	if s.Repo == "" {
		panic("Repo is empty")
	}
	if s.UnitType == "" {
		panic("UnitType is empty")
	}
	if s.Unit == "" {
		panic("Unit is empty")
	}
	return graph.DefKey{
		Repo:     repo.URI(s.Repo),
		CommitID: s.CommitID,
		UnitType: s.UnitType,
		Unit:     s.Unit,
		Path:     graph.DefPath(s.Path),
	}
}

// NewDefSpecFromDefKey returns a DefSpec that specifies the same
// def as the given key.
func NewDefSpecFromDefKey(key graph.DefKey) DefSpec {
	return DefSpec{
		Repo:     string(key.Repo),
		CommitID: key.CommitID,
		UnitType: key.UnitType,
		Unit:     key.Unit,
		Path:     string(key.Path),
	}
}

// defsService implements DefsService.
type defsService struct {
	client *Client
}

var _ DefsService = &defsService{}

// Def is a code def returned by the Sourcegraph API.
type Def struct {
	graph.Def

	Stat graph.Stats `json:",omitempty"`

	DocHTML  string           `json:",omitempty"`
	DocPages []*graph.DocPage `json:",omitempty"`
}

// DefSpec returns the DefSpec that specifies s.
func (s *Def) DefSpec() DefSpec {
	spec := NewDefSpecFromDefKey(s.Def.DefKey)
	return spec
}

func (s *Def) XRefs() int { return s.Stat["xrefs"] }
func (s *Def) RRefs() int { return s.Stat["rrefs"] }
func (s *Def) URefs() int { return s.Stat["urefs"] }

// TotalRefs is the number of unique references of all kinds to s. It
// is computed as (xrefs + rrefs), omitting urefs to avoid double-counting
// references in the same repository.
//
// The number of examples for s is usually TotalRefs() - 1, since the definition
// of a def counts as a ref but not an example.
func (s *Def) TotalRefs() int { return s.XRefs() + s.RRefs() }

func (s *Def) TotalExamples() int { return s.TotalRefs() - 1 }

// DefGetOptions specifies options for DefsService.Get.
type DefGetOptions struct {
	Doc      bool `url:",omitempty"`
	DocPages bool `url:",omitempty"`
}

func (s *defsService) Get(def DefSpec, opt *DefGetOptions) (*Def, Response, error) {
	url, err := s.client.url(router.Def, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var def_ *Def
	resp, err := s.client.Do(req, &def_)
	if err != nil {
		return nil, resp, err
	}

	return def_, resp, nil
}

// DefListOptions specifies options for DefsService.List.
type DefListOptions struct {
	Name string `url:",omitempty" json:",omitempty"`

	// Specifies a search query for defs. If specified, then the Sort and Direction options are ignored
	Query string `url:",omitempty" json:",omitempty"`

	// Filters
	RepositoryURI string   `url:",omitempty" json:",omitempty"`
	CommitID      string   `url:",omitempty" json:",omitempty"`
	UnitTypes     []string `url:",omitempty,comma" json:",omitempty"`
	Unit          string   `url:",omitempty" json:",omitempty"`

	Path string `url:",omitempty" json:",omitempty"`

	// If specified, will filter on descendants of ParentPath (up to ChildDepth)
	ParentTreePath string `url:",omitempty" json:",omitempty"`
	ChildDepth     int    `url:",omitempty" json:",omitempty"`

	// If specified, will filter on ancestors of ChildPath
	ChildTreePath string `url:",omitempty" json:",omitempty"`

	// File, if specified, will restrict the results to only defs defined in
	// the specified file.
	File string `url:",omitempty" json:",omitempty"`

	// FilePathPrefix, if specified, will restrict the results to only defs defined in
	// files whose path is underneath the specified prefix.
	FilePathPrefix string `url:",omitempty" json:",omitempty"`

	Kinds    []string `url:",omitempty,comma" json:",omitempty"`
	Exported bool     `url:",omitempty" json:",omitempty"`

	// IncludeTest is whether the results should include definitions in test
	// files.
	IncludeTest bool `url:",omitempty" json:",omitempty"`

	// Enhancements
	Doc   bool `url:",omitempty" json:",omitempty"`
	Stats bool `url:",omitempty" json:",omitempty"`

	// Sorting
	Sort      string `url:",omitempty" json:",omitempty"`
	Direction string `url:",omitempty" json:",omitempty"`

	// Paging
	ListOptions
}

func (s *defsService) List(opt *DefListOptions) ([]*Def, Response, error) {
	url, err := s.client.url(router.Defs, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var defs []*Def
	resp, err := s.client.Do(req, &defs)
	if err != nil {
		return nil, resp, err
	}

	return defs, resp, nil
}

// Example is a usage example of a def.
type Example struct {
	graph.Ref

	// SrcHTML is the formatted HTML source code of the example, with links to
	// definitions.
	SrcHTML template.HTML

	// The line that the given example starts on
	StartLine int

	// The line that the given example ends on
	EndLine int

	// Error is whether an error occurred while fetching this example.
	Error bool
}

type Examples []*Example

func (r *Example) sortKey() string     { return fmt.Sprintf("%+v", r) }
func (vs Examples) Len() int           { return len(vs) }
func (vs Examples) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Examples) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }

// DefListExamplesOptions specifies options for DefsService.ListExamples.
type DefListExamplesOptions struct {
	Formatted bool

	// Filter by a specific Repository URI
	Repository string

	// Filter by a specific User
	User string

	ListOptions
}

func (s *defsService) ListExamples(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error) {
	url, err := s.client.url(router.DefExamples, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var examples []*Example
	resp, err := s.client.Do(req, &examples)
	if err != nil {
		return nil, resp, err
	}

	return examples, resp, nil
}

type AuthorshipInfo struct {
	AuthorEmail    string    `db:"author_email"`
	LastCommitDate time.Time `db:"last_commit_date"`

	// LastCommitID is the commit ID of the last commit that this author made to
	// the thing that this info describes.
	LastCommitID string `db:"last_commit_id"`
}

type DefAuthorship struct {
	AuthorshipInfo

	// Exported is whether the def is exported.
	Exported bool

	Bytes           int
	BytesProportion float64
}

type DefAuthor struct {
	UID   nnz.Int
	Email nnz.String
	DefAuthorship
}

type DefAuthorsByBytes []*DefAuthor

func (v DefAuthorsByBytes) Len() int           { return len(v) }
func (v DefAuthorsByBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v DefAuthorsByBytes) Less(i, j int) bool { return v[i].Bytes < v[j].Bytes }

type AugmentedDefAuthor struct {
	User *person.User
	*DefAuthor
}

// DefListAuthorsOptions specifies options for DefsService.ListAuthors.
type DefListAuthorsOptions struct {
	ListOptions
}

func (s *defsService) ListAuthors(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error) {
	url, err := s.client.url(router.DefAuthors, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var authors []*AugmentedDefAuthor
	resp, err := s.client.Do(req, &authors)
	if err != nil {
		return nil, resp, err
	}

	return authors, resp, nil
}

// RefAuthorship describes the authorship information (author email, date, and
// commit ID) of a ref. A ref may only have one author.
type RefAuthorship struct {
	graph.RefKey
	AuthorshipInfo
}

type DefClient struct {
	UID   nnz.Int
	Email nnz.String

	AuthorshipInfo

	// UseCount is the number of times this person referred to the def.
	UseCount int `db:"use_count"`
}

type AugmentedDefClient struct {
	User *person.User
	*DefClient
}

// DefListClientsOptions specifies options for DefsService.ListClients.
type DefListClientsOptions struct {
	ListOptions
}

func (s *defsService) ListClients(def DefSpec, opt *DefListClientsOptions) ([]*AugmentedDefClient, Response, error) {
	url, err := s.client.url(router.DefClients, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var clients []*AugmentedDefClient
	resp, err := s.client.Do(req, &clients)
	if err != nil {
		return nil, resp, err
	}

	return clients, resp, nil
}

type DefDependent struct {
	FromRepo repo.URI `db:"from_repo"`
	Count    int
}

type AugmentedDefDependent struct {
	Repo *repo.Repository
	*DefDependent
}

// DefListDependentsOptions specifies options for DefsService.ListDependents.
type DefListDependentsOptions struct {
	ListOptions
}

func (s *defsService) ListDependents(def DefSpec, opt *DefListDependentsOptions) ([]*AugmentedDefDependent, Response, error) {
	url, err := s.client.url(router.DefDependents, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var dependents []*AugmentedDefDependent
	resp, err := s.client.Do(req, &dependents)
	if err != nil {
		return nil, resp, err
	}

	return dependents, resp, nil
}

type DefListVersionsOptions struct {
	ListOptions
}

func (s *defsService) ListVersions(def DefSpec, opt *DefListVersionsOptions) ([]*Def, Response, error) {
	url, err := s.client.url(router.DefVersions, def.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var defVersions []*Def
	resp, err := s.client.Do(req, &defVersions)
	if err != nil {
		return nil, resp, err
	}

	return defVersions, resp, nil
}

type MockDefsService struct {
	Get_            func(def DefSpec, opt *DefGetOptions) (*Def, Response, error)
	List_           func(opt *DefListOptions) ([]*Def, Response, error)
	ListExamples_   func(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error)
	ListAuthors_    func(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error)
	ListClients_    func(def DefSpec, opt *DefListClientsOptions) ([]*AugmentedDefClient, Response, error)
	ListDependents_ func(def DefSpec, opt *DefListDependentsOptions) ([]*AugmentedDefDependent, Response, error)
	ListVersions_   func(def DefSpec, opt *DefListVersionsOptions) ([]*Def, Response, error)
}

var _ DefsService = MockDefsService{}

func (s MockDefsService) Get(def DefSpec, opt *DefGetOptions) (*Def, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(def, opt)
}

func (s MockDefsService) List(opt *DefListOptions) ([]*Def, Response, error) {
	if s.List_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.List_(opt)
}

func (s MockDefsService) ListExamples(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error) {
	if s.ListExamples_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListExamples_(def, opt)
}

func (s MockDefsService) ListAuthors(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error) {
	if s.ListAuthors_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListAuthors_(def, opt)
}

func (s MockDefsService) ListClients(def DefSpec, opt *DefListClientsOptions) ([]*AugmentedDefClient, Response, error) {
	if s.ListClients_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListClients_(def, opt)
}

func (s MockDefsService) ListDependents(def DefSpec, opt *DefListDependentsOptions) ([]*AugmentedDefDependent, Response, error) {
	if s.ListDependents_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListDependents_(def, opt)
}

func (s MockDefsService) ListVersions(def DefSpec, opt *DefListVersionsOptions) ([]*Def, Response, error) {
	if s.ListVersions_ == nil {
		return nil, nil, nil
	}
	return s.ListVersions_(def, opt)
}
