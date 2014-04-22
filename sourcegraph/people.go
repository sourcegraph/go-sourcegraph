package client

import (
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/api_router"
	"sourcegraph.com/sourcegraph/srcgraph/person"
)

type PeopleService interface {
	Get(person PersonSpec) (*person.User, *Response, error)
	List(opt *PeopleListOptions) ([]*person.User, *Response, error)
	ListAuthors(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error)
	ListClients(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error)
}

type peopleService struct {
	client *Client
}

var _ PeopleService = &peopleService{}

type PersonSpec struct {
	Email string
	Login string
	UID   int
}

func (s *PersonSpec) PathComponent() string {
	if s.Email != "" {
		return s.Email
	}
	if s.Login != "" {
		return s.Login
	}
	if s.UID > 0 {
		return "$" + strconv.Itoa(s.UID)
	}
	panic("empty PersonSpec")
}

// ParsePersonSpec parses a string generated by (*PersonSpec).String() and
// returns the equivalent PersonSpec struct.
func ParsePersonSpec(pathComponent string) (PersonSpec, error) {
	if strings.HasPrefix(pathComponent, "$") {
		uid, err := strconv.Atoi(pathComponent[1:])
		return PersonSpec{UID: uid}, err
	}
	if strings.Contains(pathComponent, "@") {
		return PersonSpec{Email: pathComponent}, nil
	}
	return PersonSpec{Login: pathComponent}, nil
}

func (s *peopleService) Get(person_ PersonSpec) (*person.User, *Response, error) {
	url, err := s.client.url(api_router.Person, map[string]string{"PersonSpec": person_.PathComponent()}, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var person__ *person.User
	resp, err := s.client.Do(req, &person__)
	if err != nil {
		return nil, resp, err
	}

	return person__, resp, nil
}

type PeopleListOptions struct {
	Query string `url:",omitempty"`

	Sort      string `url:",omitempty"`
	Direction string `url:",omitempty"`

	ListOptions
}

func (s *peopleService) List(opt *PeopleListOptions) ([]*person.User, *Response, error) {
	url, err := s.client.url(api_router.People, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var people []*person.User
	resp, err := s.client.Do(req, &people)
	if err != nil {
		return nil, resp, err
	}

	return people, resp, nil
}

// AugmentedPersonRef is a rel.PersonRef with the full person.User struct embedded.
type AugmentedPersonRef struct {
	User  *person.User `json:"user"`
	Count int          `json:"count"`
}

func (s *peopleService) listPersonPersonRefs(person PersonSpec, routeName string, opt interface{}) ([]*AugmentedPersonRef, *Response, error) {
	url, err := s.client.url(routeName, map[string]string{"PersonSpec": person.PathComponent()}, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var people []*AugmentedPersonRef
	resp, err := s.client.Do(req, &people)
	if err != nil {
		return nil, resp, err
	}

	return people, resp, nil
}

func (s *peopleService) ListAuthors(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error) {
	return s.listPersonPersonRefs(person, api_router.PersonAuthors, opt)
}

func (s *peopleService) ListClients(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error) {
	return s.listPersonPersonRefs(person, api_router.PersonClients, opt)
}

type MockPeopleService struct {
	Get_         func(person PersonSpec) (*person.User, *Response, error)
	List_        func(opt *PeopleListOptions) ([]*person.User, *Response, error)
	ListAuthors_ func(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error)
	ListClients_ func(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error)
}

var _ PeopleService = MockPeopleService{}

func (s MockPeopleService) Get(person PersonSpec) (*person.User, *Response, error) {
	if s.Get_ == nil {
		return nil, &Response{}, nil
	}
	return s.Get_(person)
}

func (s MockPeopleService) List(opt *PeopleListOptions) ([]*person.User, *Response, error) {
	if s.List_ == nil {
		return nil, &Response{}, nil
	}
	return s.List_(opt)
}

func (s MockPeopleService) ListAuthors(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error) {
	if s.ListAuthors_ == nil {
		return nil, &Response{}, nil
	}
	return s.ListAuthors_(person, opt)
}

func (s MockPeopleService) ListClients(person PersonSpec, opt *PeopleListOptions) ([]*AugmentedPersonRef, *Response, error) {
	if s.ListClients_ == nil {
		return nil, &Response{}, nil
	}
	return s.ListClients_(person, opt)
}
