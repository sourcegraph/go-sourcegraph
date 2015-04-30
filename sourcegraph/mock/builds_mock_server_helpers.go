package mock

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *BuildsServer) MockGetRepoBuildInfo(t *testing.T, info *sourcegraph.RepoBuildInfo) (called *bool) {
	called = new(bool)
	s.GetRepoBuildInfo_ = func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		*called = true
		return info, nil
	}
	return
}
