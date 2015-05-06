package mock

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *UsersServer) MockGet(t *testing.T, wantUser string) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.Login != wantUser {
			t.Errorf("got user %q, want %q", user.Login, wantUser)
			return nil, sourcegraph.ErrNotExist
		}
		return &sourcegraph.User{Login: user.Login}, nil
	}
	return
}

func (s *UsersServer) MockGet_Return(t *testing.T, returns *sourcegraph.User) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.Login != returns.Login {
			t.Errorf("got user %q, want %q", user.Login, returns.Login)
			return nil, sourcegraph.ErrNotExist
		}
		return returns, nil
	}
	return
}

func (s *UsersServer) MockList(t *testing.T, wantUsers ...string) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
		*called = true
		users := make([]*sourcegraph.User, len(wantUsers))
		for i, user := range wantUsers {
			users[i] = &sourcegraph.User{Login: user}
		}
		return &sourcegraph.UserList{Users: users}, nil
	}
	return
}
