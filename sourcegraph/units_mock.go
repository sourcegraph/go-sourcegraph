package sourcegraph

import "sourcegraph.com/sourcegraph/srclib/unit"

type MockUnitsService struct {
	Get_  func(spec UnitSpec) (*unit.RepoSourceUnit, Response, error)
	List_ func(opt *UnitListOptions) ([]*unit.RepoSourceUnit, Response, error)
}

func (s MockUnitsService) Get(spec UnitSpec) (*unit.RepoSourceUnit, Response, error) {
	return s.Get_(spec)
}

func (s MockUnitsService) List(opt *UnitListOptions) ([]*unit.RepoSourceUnit, Response, error) {
	return s.List_(opt)
}
