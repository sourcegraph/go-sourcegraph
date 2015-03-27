package sourcegraph

type MockNotesService struct {
	List_   func(opt *NotesListOptions) ([]*Note, Response, error)
	Create_ func(note *Note) (*Note, Response, error)
}

func (s MockNotesService) List(opt *NotesListOptions) ([]*Note, Response, error) {
	return s.List_(opt)
}

func (s MockNotesService) Create(note *Note) (*Note, Response, error) {
	return s.Create_(note)
}
