package sourcegraph

type MockDefsService struct {
	Get_          func(def DefSpec, opt *DefGetOptions) (*Def, Response, error)
	List_         func(opt *DefListOptions) ([]*Def, Response, error)
	ListRefs_     func(def DefSpec, opt *DefListRefsOptions) ([]*Ref, Response, error)
	ListExamples_ func(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error)
	ListAuthors_  func(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error)
}

func (s MockDefsService) Get(def DefSpec, opt *DefGetOptions) (*Def, Response, error) {
	return s.Get_(def, opt)
}

func (s MockDefsService) List(opt *DefListOptions) ([]*Def, Response, error) { return s.List_(opt) }

func (s MockDefsService) ListRefs(def DefSpec, opt *DefListRefsOptions) ([]*Ref, Response, error) {
	return s.ListRefs_(def, opt)
}

func (s MockDefsService) ListExamples(def DefSpec, opt *DefListExamplesOptions) ([]*Example, Response, error) {
	return s.ListExamples_(def, opt)
}

func (s MockDefsService) ListAuthors(def DefSpec, opt *DefListAuthorsOptions) ([]*AugmentedDefAuthor, Response, error) {
	return s.ListAuthors_(def, opt)
}
