package sourcegraph

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

var (
	baseRev = RepoRevSpec{RepoSpec: RepoSpec{URI: "base.com/repo"}, Rev: "baserev", CommitID: "basecommit"}
	headRev = RepoRevSpec{RepoSpec: RepoSpec{URI: "head.com/repo"}, Rev: "headrev", CommitID: "headcommit"}
)

func TestDeltas(t *testing.T) {
	tests := []struct {
		spec          DeltaSpec
		wantRouteVars map[string]string
	}{
		{
			spec: DeltaSpec{
				Base: RepoRevSpec{RepoSpec: RepoSpec{URI: "samerepo"}, Rev: "baserev", CommitID: "basecommit"},
				Head: RepoRevSpec{RepoSpec: RepoSpec{URI: "samerepo"}, Rev: "headrev", CommitID: "headcommit"},
			},
			wantRouteVars: map[string]string{
				"RepoSpec":     "samerepo",
				"Rev":          "baserev===basecommit",
				"DeltaHeadRev": "headrev===headcommit",
			},
		},
		{
			spec: DeltaSpec{
				Base: baseRev,
				Head: headRev,
			},
			wantRouteVars: map[string]string{
				"RepoSpec":     "base.com/repo",
				"Rev":          "baserev===basecommit",
				"DeltaHeadRev": encodeCrossRepoRevSpecForDeltaHeadRev(headRev),
			},
		},
	}
	for _, test := range tests {
		vars := test.spec.RouteVars()
		if !reflect.DeepEqual(vars, test.wantRouteVars) {
			t.Errorf("got route vars != want\n\n%s", strings.Join(pretty.Diff(vars, test.wantRouteVars), "\n"))
		}

		spec, err := UnmarshalDeltaSpec(vars)
		if err != nil {
			t.Errorf("UnmarshalDeltaSpec(%+v): %s", err)
			continue
		}
		if !reflect.DeepEqual(spec, test.spec) {
			t.Errorf("got spec != original spec\n\n%s", strings.Join(pretty.Diff(spec, test.spec), "\n"))
		}
	}
}
