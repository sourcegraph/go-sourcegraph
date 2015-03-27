package sourcegraph

import (
	"time"

	"strconv"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// NotesService communicates with the note-related endpoints in the
// Sourcegraph API.
type NotesService interface {
	// List lists notes that match the specified options.
	List(opt *NotesListOptions) ([]*Note, Response, error)

	// Create creates a new note.
	Create(note *Note) (*Note, Response, error)
}

// notesService implements ReposService.
type notesService struct {
	client *Client
}

var _ NotesService = &notesService{}

// NoteSpec specifies a note.
type NoteSpec struct {
	ID int
}

// RouteVars returns route variables for constructing note routes.
func (s NoteSpec) RouteVars() map[string]string {
	return map[string]string{"NoteSpec": strconv.Itoa(s.ID)}
}

// UnmarshalNoteSpec parses and returns the NoteSpec present in the
// route variables map. If no NoteSpec exists in the map, nil is
// returned.
func UnmarshalNoteSpec(routeVars map[string]string) *NoteSpec {
	if idStr, ok := routeVars["NoteSpec"]; ok {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			return &NoteSpec{ID: id}
		}
	}
	return nil
}

// A Note is a question, comment, etc., about code.
type Note struct {
	// ID is the unique ID of the note.
	ID int

	// Obj is the object the note is attached to.
	Obj NoteObject

	// Author is the author of the note.
	Author UserSpec

	// Body is the content of the note (which may contain Markdown
	// formatting).
	Body string

	// CreatedAt is the time that this note was created by the author.
	CreatedAt time.Time

	// UpdatedAt is the last modification time of this note. It is nil
	// if the note has not been modified since creation.
	UpdatedAt *time.Time

	// ClosedBy is the user who closed this note (if any).
	ClosedBy *UserSpec

	// ClosedAt is the time at which this note was closed (if any).
	ClosedAt *time.Time

	// Deleted is whether this note was deleted (by the author or an
	// admin).
	Deleted bool `json:",omitempty"`
}

// A NoteObject is the object that a Note is attached to: a def, an
// offset within a def, an offset within a file, a directory, a
// commit, a tag/branch, a repository, or another note (in the case of
// a note reply).
type NoteObject struct {
	// ParentNoteID is the ID of this note's parent (if any).
	ParentNoteID int `json:",omitempty"`

	// Repo is the repository that this note is attached to.
	Repo *RepoSpec

	// Rev is the commit or revision that this note is attached to (if
	// any).
	Rev string

	// Def is the definition that this note is attached to (if any).
	Def *DefSpec

	// Entry is the file or directory that this note is attached to
	// (if any).
	TreeEntry *TreeEntrySpec

	// LineStart and LineEnd are the 1-indexed line range that this
	// note is attached to. If this is a file note, these are the
	// lines in the file (subject to shifting based on file diff
	// changes since the note was made). If this is a def note, this
	// are the lines relative to the line containing the DefStart of
	// the def.
	LineStart, LineEnd int64
}

// NotesListOptions specifies options to Notes.List.
type NotesListOptions struct {
	Sort      string `url:",omitempty"`
	Direction string `url:",omitempty"`

	Repo string `url:",omitempty"`
	Rev  string `url:",omitempty"`

	UnitType string `url:",omitempty"`
	Unit     string `url:",omitempty"`
	DefPath  string `url:",omitempty"`

	TreeEntryPath string `url:",omitempty"`

	LineStart int64 `url:",omitempty"`
	LineEnd   int64 `url:",omitempty"`

	ParentNote int `url:",omitempty"`

	ListOptions
}

// Select returns a boolean value indicating whether o's filter
// options (i.e., without respect to sorting or pagination) would
// select note.
func (o *NotesListOptions) Select(note *Note) bool {
	switch {
	case note.Obj.ParentNoteID != 0:
		return o.ParentNote == note.Obj.ParentNoteID
	case note.Obj.Repo != nil:
		return graph.URIEqual(o.Repo, note.Obj.Repo.URI) && o.Rev == note.Obj.Rev
	case note.Obj.Def != nil:
		// Ignore commit ID. TODO(sqs): how to handle this?
		def := *note.Obj.Def
		def.CommitID = ""
		return o.Repo == def.Repo && o.UnitType == def.UnitType && o.Unit == def.Unit && o.DefPath == def.Path
	case note.Obj.TreeEntry != nil:
		// Ignore commit ID. TODO(sqs): how to handle this?
		e := *note.Obj.TreeEntry
		e.RepoRev.Rev = ""
		e.RepoRev.CommitID = ""
		lineMatch := (o.LineStart == 0 && o.LineEnd == 0) || (note.Obj.LineStart != 0 && note.Obj.LineEnd != 0 && o.LineStart >= note.Obj.LineStart && o.LineEnd <= note.Obj.LineEnd)
		return o.Repo == e.RepoRev.URI && o.TreeEntryPath == e.Path && lineMatch
	}
	return false
}

func (s *notesService) List(opt *NotesListOptions) ([]*Note, Response, error) {
	url, err := s.client.URL(router.Notes, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var notes []*Note
	resp, err := s.client.Do(req, &notes)
	if err != nil {
		return nil, resp, err
	}

	return notes, resp, nil
}

func (s *notesService) Create(note *Note) (*Note, Response, error) {
	url, err := s.client.URL(router.NotesCreate, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", url.String(), note)
	if err != nil {
		return nil, nil, err
	}

	var created Note
	resp, err := s.client.Do(req, &created)
	if err != nil {
		return nil, resp, err
	}

	return &created, resp, nil
}
