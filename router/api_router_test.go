package router

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/sqs/mux"
)

func TestMatch(t *testing.T) {
	router := NewAPIRouter("/")
	tests := []struct {
		path          string
		wantNoMatch   bool
		wantRouteName string
		wantVars      map[string]string
		wantPath      string
	}{
		// Repository
		{
			path:          "/repos/repohost.com/foo",
			wantRouteName: Repository,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo"},
		},
		{
			path:          "/repos/a/b/c",
			wantRouteName: Repository,
			wantVars:      map[string]string{"RepoSpec": "a/b/c"},
		},
		{
			path:          "/repos/a.com/b",
			wantRouteName: Repository,
			wantVars:      map[string]string{"RepoSpec": "a.com/b"},
		},
		{
			path:          "/repos/a.com/b@mycommitid",
			wantRouteName: Repository,
			wantVars:      map[string]string{"RepoSpec": "a.com/b", "Rev": "mycommitid"},
		},
		{
			path:        "/repos/.invalidrepo",
			wantNoMatch: true,
		},

		// Repository sub-routes
		{
			path:          "/repos/repohost.com/foo/.authors",
			wantRouteName: RepositoryAuthors,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo"},
		},
		{
			path:          "/repos/repohost.com/foo@myrev/.authors",
			wantRouteName: RepositoryAuthors,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "myrev"},
		},

		// Repository sub-routes that don't allow an "@REVSPEC" revision.
		{
			path:          "/repos/repohost.com/foo/.dependents",
			wantRouteName: RepositoryDependents,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo"},
		},
		{
			path:        "/repos/repohost.com/foo@myrevspec/.dependents", // no @REVSPEC match
			wantNoMatch: true,
		},
		{
			path:          "/repos/repohost.com/foo/.commits",
			wantRouteName: RepoCommits,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo"},
		},
		{
			path:          "/repos/repohost.com/foo/.commits/123abc",
			wantRouteName: RepoCommit,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "123abc"},
		},
		{
			path:          "/repos/repohost.com/foo/.commits/123abc/compare",
			wantRouteName: RepoCompareCommits,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "123abc"},
		},

		// Repository tree
		{
			path:          "/repos/repohost.com/foo@mycommitid/.tree",
			wantRouteName: RepositoryTreeEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "mycommitid", "Path": "."},
		},
		{
			path:          "/repos/repohost.com/foo@mycommitid/.tree/",
			wantRouteName: RepositoryTreeEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "mycommitid", "Path": "."},
			wantPath:      "/repos/repohost.com/foo@mycommitid/.tree",
		},
		{
			path:          "/repos/repohost.com/foo@mycommitid/.tree/my/file",
			wantRouteName: RepositoryTreeEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "mycommitid", "Path": "my/file"},
		},

		// Repository build data
		{
			path:          "/repos/repohost.com/foo/.build-data",
			wantRouteName: RepositoryBuildDataEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Path": "."},
		},
		{
			path:          "/repos/repohost.com/foo@mycommitid/.build-data/",
			wantRouteName: RepositoryBuildDataEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "mycommitid", "Path": "."},
			wantPath:      "/repos/repohost.com/foo@mycommitid/.build-data",
		},
		{
			path:          "/repos/repohost.com/foo@mycommitid/.build-data/my/file",
			wantRouteName: RepositoryBuildDataEntry,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "Rev": "mycommitid", "Path": "my/file"},
		},

		// Defs
		{
			path:          "/repos/repohost.com/foo@mycommitid/.defs/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p", "Rev": "mycommitid"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def", // empty path
			wantRouteName: Def,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "."},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/u1/.def/p",
			wantRouteName: Def,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": "u1", "Path": "p"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/u1/u2/.def/p1/p2",
			wantRouteName: Def,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": "u1/u2", "Path": "p1/p2"},
		},

		// Def sub-routes
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def/p/.authors",
			wantRouteName: DefAuthors,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "p"},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/.def/.authors", // empty path
			wantRouteName: DefAuthors,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": ".", "Path": "."},
		},
		{
			path:          "/repos/repohost.com/foo/.defs/.t/u1/u2/.def/p1/p2/.authors",
			wantRouteName: DefAuthors,
			wantVars:      map[string]string{"RepoSpec": "repohost.com/foo", "UnitType": "t", "Unit": "u1/u2", "Path": "p1/p2"},
		},

		// Person
		{
			path:          "/people/alice",
			wantRouteName: Person,
			wantVars:      map[string]string{"PersonSpec": "alice"},
		},
		{
			path:          "/people/alice@example.com",
			wantRouteName: Person,
			wantVars:      map[string]string{"PersonSpec": "alice@example.com"},
		},
		{
			path:          "/people/alice@-x-yJAANTud-iAVVw==",
			wantRouteName: Person,
			wantVars:      map[string]string{"PersonSpec": "alice@-x-yJAANTud-iAVVw=="},
		},
	}
	for _, test := range tests {
		var routeMatch mux.RouteMatch
		match := router.Match(&http.Request{Method: "GET", URL: &url.URL{Path: test.path}}, &routeMatch)

		if match && test.wantNoMatch {
			t.Errorf("%s: got match (route %q), want no match", test.path, routeMatch.Route.GetName())
		}
		if !match && !test.wantNoMatch {
			t.Errorf("%s: got no match, wanted match", test.path)
		}
		if !match || test.wantNoMatch {
			continue
		}

		if routeName := routeMatch.Route.GetName(); routeName != test.wantRouteName {
			t.Errorf("%s: got matched route %q, want %q", test.path, routeName, test.wantRouteName)
		}

		if diff := pretty.Diff(routeMatch.Vars, test.wantVars); len(diff) > 0 {
			t.Errorf("%s: vars don't match expected:\n%s", test.path, strings.Join(diff, "\n"))
		}

		// Check that building the URL yields the original path.
		var pairs []string
		for k, v := range test.wantVars {
			pairs = append(pairs, k, v)
		}
		path, err := routeMatch.Route.URLPath(pairs...)
		if err != nil {
			t.Errorf("%s: URLPath(%v) failed: %s", test.path, pairs, err)
			continue
		}
		var wantPath string
		if test.wantPath != "" {
			wantPath = test.wantPath
		} else {
			wantPath = test.path
		}
		if path.Path != wantPath {
			t.Errorf("got generated path %q, want %q", path, wantPath)
		}
	}
}
