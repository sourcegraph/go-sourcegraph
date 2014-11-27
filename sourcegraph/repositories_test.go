package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"strings"

	"github.com/kr/pretty"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/person"
	"sourcegraph.com/sourcegraph/srclib/repo"
)

func TestRepoSpec(t *testing.T) {
	tests := []struct {
		str  string
		spec RepoSpec
	}{
		{"a.com/x", RepoSpec{URI: "a.com/x"}},
		{"R$1", RepoSpec{RID: 1}},
	}

	for _, test := range tests {
		spec, err := ParseRepoSpec(test.str)
		if err != nil {
			t.Errorf("%q: ParseRepoSpec failed: %s", test.str, err)
			continue
		}
		if spec != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}

		str := test.spec.PathComponent()
		if str != test.str {
			t.Errorf("%+v: got str %q, want %q", test.spec, str, test.str)
			continue
		}

		spec2, err := UnmarshalRepoSpec(test.spec.RouteVars())
		if err != nil {
			t.Errorf("%+v: UnmarshalRepoSpec: %s", test.spec, err)
			continue
		}
		if spec2 != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}
	}
}

func TestRepoRevSpec(t *testing.T) {
	tests := []struct {
		spec      RepoRevSpec
		routeVars map[string]string
	}{
		{RepoRevSpec{RepoSpec: RepoSpec{URI: "a.com/x"}, Rev: "r"}, map[string]string{"RepoSpec": "a.com/x", "Rev": "r"}},
		{RepoRevSpec{RepoSpec: RepoSpec{RID: 123}, Rev: "r"}, map[string]string{"RepoSpec": "R$123", "Rev": "r"}},
		{RepoRevSpec{RepoSpec: RepoSpec{URI: "a.com/x"}, Rev: "r", CommitID: "c"}, map[string]string{"RepoSpec": "a.com/x", "Rev": "r===c"}},
	}

	for _, test := range tests {
		routeVars := test.spec.RouteVars()
		if !reflect.DeepEqual(routeVars, test.routeVars) {
			t.Errorf("got route vars %+v, want %+v", routeVars, test.routeVars)
		}
		spec, err := UnmarshalRepoRevSpec(routeVars)
		if err != nil {
			t.Errorf("UnmarshalRepoRevSpec(%+v): %s", routeVars, err)
			continue
		}
		if spec != test.spec {
			t.Errorf("got spec %+v, want %+v", spec, test.spec)
		}
	}
}

func TestRepositoriesService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &Repository{Repository: &repo.Repository{RID: 1}}

	var called bool
	mux.HandleFunc(urlPath(t, router.Repository, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	repo_, _, err := client.Repositories.Get(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repo_, want) {
		t.Errorf("Repositories.Get returned %+v, want %+v", repo_, want)
	}
}

func TestRepositoriesService_GetStats(t *testing.T) {
	setup()
	defer teardown()

	want := repo.Stats{"x": 1, "y": 2}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryStats, map[string]string{"RepoSpec": "r.com/x", "Rev": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	stats, _, err := client.Repositories.GetStats(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"})
	if err != nil {
		t.Errorf("Repositories.GetStats returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(stats, want) {
		t.Errorf("Repositories.GetStats returned %+v, want %+v", stats, want)
	}
}

func TestRepositoriesService_GetOrCreate(t *testing.T) {
	setup()
	defer teardown()

	want := &Repository{Repository: &repo.Repository{RID: 1}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoriesGetOrCreate, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")

		writeJSON(w, want)
	})

	repo_, _, err := client.Repositories.GetOrCreate(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.GetOrCreate returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repo_, want) {
		t.Errorf("Repositories.GetOrCreate returned %+v, want %+v", repo_, want)
	}
}

func TestRepositoriesService_GetSettings(t *testing.T) {
	setup()
	defer teardown()

	want := &RepositorySettings{}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositorySettings, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	settings, _, err := client.Repositories.GetSettings(RepoSpec{URI: "r.com/x"})
	if err != nil {
		t.Errorf("Repositories.GetSettings returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Repositories.GetSettings returned %+v, want %+v", settings, want)
	}
}

func TestRepositoriesService_UpdateSettings(t *testing.T) {
	setup()
	defer teardown()

	want := RepositorySettings{}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositorySettings, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
		testBody(t, r, `{}`+"\n")

		writeJSON(w, want)
	})

	_, err := client.Repositories.UpdateSettings(RepoSpec{URI: "r.com/x"}, want)
	if err != nil {
		t.Errorf("Repositories.UpdateSettings returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestRepositoriesService_RefreshProfile(t *testing.T) {
	setup()
	defer teardown()

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryRefreshProfile, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
	})

	_, err := client.Repositories.RefreshProfile(RepoSpec{URI: "r.com/x"})
	if err != nil {
		t.Errorf("Repositories.RefreshProfile returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestRepositoriesService_RefreshVCSData(t *testing.T) {
	setup()
	defer teardown()

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryRefreshVCSData, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
	})

	_, err := client.Repositories.RefreshVCSData(RepoSpec{URI: "r.com/x"})
	if err != nil {
		t.Errorf("Repositories.RefreshVCSData returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestRepositoriesService_ComputeStats(t *testing.T) {
	setup()
	defer teardown()

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryComputeStats, map[string]string{"RepoSpec": "r.com/x", "Rev": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
	})

	_, err := client.Repositories.ComputeStats(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"})
	if err != nil {
		t.Errorf("Repositories.ComputeStats returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestRepositoriesService_GetBuild(t *testing.T) {
	setup()
	defer teardown()

	want := &RepoBuildInfo{
		Exact:                &Build{BID: 1},
		LastSuccessful:       &Build{BID: 2},
		CommitsBehind:        3,
		LastSuccessfulCommit: &Commit{Commit: &vcs.Commit{Message: "m"}},
	}
	normalizeTime(&want.LastSuccessfulCommit.Author.Date)
	normalizeBuildTime(want.Exact)
	normalizeBuildTime(want.LastSuccessful)

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBuild, map[string]string{"RepoSpec": "r.com/x", "Rev": "r"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	buildInfo, _, err := client.Repositories.GetBuild(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "r"}, nil)
	if err != nil {
		t.Errorf("Repositories.GetBuild returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeTime(&buildInfo.LastSuccessfulCommit.Author.Date)
	normalizeBuildTime(buildInfo.Exact)
	normalizeBuildTime(buildInfo.LastSuccessful)
	if !reflect.DeepEqual(buildInfo.Exact, want.Exact) {
		t.Errorf("Repositories.GetBuild returned %+v, want %+v", buildInfo.Exact, want.Exact)
	}
}

func TestRepositoriesService_Create(t *testing.T) {
	setup()
	defer teardown()

	newRepo := NewRepositorySpec{Type: "git", CloneURLStr: "http://r.com/x"}
	want := &repo.Repository{RID: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoriesCreate, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		testBody(t, r, `{"Type":"git","CloneURL":"http://r.com/x"}`+"\n")

		writeJSON(w, want)
	})

	repo_, _, err := client.Repositories.Create(newRepo)
	if err != nil {
		t.Errorf("Repositories.Create returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repo_, want) {
		t.Errorf("Repositories.Create returned %+v, want %+v", repo_, want)
	}
}

func TestRepositoriesService_GetReadme(t *testing.T) {
	setup()
	defer teardown()

	want := &vcsclient.TreeEntry{Name: "hello"}
	want.ModTime = want.ModTime.In(time.UTC)

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryReadme, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	readme, _, err := client.Repositories.GetReadme(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}})
	if err != nil {
		t.Errorf("Repositories.GetReadme returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(readme, want) {
		t.Errorf("Repositories.GetReadme returned %+v, want %+v", readme, want)
	}
}

func TestRepositoriesService_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*Repository{&Repository{Repository: &repo.Repository{RID: 1}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.Repositories, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"URIs":      "a,b",
			"Name":      "n",
			"Owner":     "o",
			"Sort":      "name",
			"Direction": "asc",
			"NoFork":    "true",
			"PerPage":   "1",
			"Page":      "2",
		})

		writeJSON(w, want)
	})

	repos, _, err := client.Repositories.List(&RepositoryListOptions{
		URIs:        []string{"a", "b"},
		Name:        "n",
		Owner:       "o",
		Sort:        "name",
		Direction:   "asc",
		NoFork:      true,
		ListOptions: ListOptions{PerPage: 1, Page: 2},
	})
	if err != nil {
		t.Errorf("Repositories.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.List returned %+v, want %+v with diff: %s", repos, want, strings.Join(pretty.Diff(want, repos), "\n"))
	}
}

func TestRepositoriesService_ListCommits(t *testing.T) {
	setup()
	defer teardown()

	want := []*Commit{{Commit: &vcs.Commit{Message: "m"}}}
	normTime(want[0])

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoCommits, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"Head": "myhead"})

		writeJSON(w, want)
	})

	commits, _, err := client.Repositories.ListCommits(RepoSpec{URI: "r.com/x"}, &RepositoryListCommitsOptions{Head: "myhead"})
	if err != nil {
		t.Errorf("Repositories.ListCommits returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(commits, want) {
		t.Errorf("Repositories.ListCommits returned %+v, want %+v", commits, want)
	}
}

func TestRepositoriesService_GetCommit(t *testing.T) {
	setup()
	defer teardown()

	want := &Commit{Commit: &vcs.Commit{Message: "m"}}
	normTime(want)

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoCommit, map[string]string{"RepoSpec": "r.com/x", "Rev": "r"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	commit, _, err := client.Repositories.GetCommit(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "r"}, nil)
	if err != nil {
		t.Errorf("Repositories.GetCommit returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(commit, want) {
		t.Errorf("Repositories.GetCommit returned %+v, want %+v", commit, want)
	}
}

func TestRepositoriesService_ListBranches(t *testing.T) {
	setup()
	defer teardown()

	want := []*vcs.Branch{{Name: "b", Head: "c"}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBranches, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	branches, _, err := client.Repositories.ListBranches(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListBranches returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(branches, want) {
		t.Errorf("Repositories.ListBranches returned %+v, want %+v", branches, want)
	}
}

func TestRepositoriesService_ListTags(t *testing.T) {
	setup()
	defer teardown()

	want := []*vcs.Tag{{Name: "t", CommitID: "c"}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoTags, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	tags, _, err := client.Repositories.ListTags(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListTags returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(tags, want) {
		t.Errorf("Repositories.ListTags returned %+v, want %+v", tags, want)
	}
}

func TestRepositoriesService_ListBadges(t *testing.T) {
	setup()
	defer teardown()

	want := []*Badge{{Name: "b"}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryBadges, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	badges, _, err := client.Repositories.ListBadges(RepoSpec{URI: "r.com/x"})
	if err != nil {
		t.Errorf("Repositories.ListBadges returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(badges, want) {
		t.Errorf("Repositories.ListBadges returned %+v, want %+v", badges, want)
	}
}

func TestRepositoriesService_ListCounters(t *testing.T) {
	setup()
	defer teardown()

	want := []*Counter{{Name: "b"}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryCounters, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	counters, _, err := client.Repositories.ListCounters(RepoSpec{URI: "r.com/x"})
	if err != nil {
		t.Errorf("Repositories.ListCounters returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(counters, want) {
		t.Errorf("Repositories.ListCounters returned %+v, want %+v", counters, want)
	}
}

func TestRepositoriesService_ListAuthors(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoAuthor{{User: &person.User{Login: "b"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryAuthors, map[string]string{"RepoSpec": "r.com/x", "Rev": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	authors, _, err := client.Repositories.ListAuthors(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListAuthors returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(authors, want) {
		t.Errorf("Repositories.ListAuthors returned %+v, want %+v", authors, want)
	}
}

func TestRepositoriesService_ListClients(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoClient{{User: &person.User{Login: "b"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryClients, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	clients, _, err := client.Repositories.ListClients(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListClients returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(clients, want) {
		t.Errorf("Repositories.ListClients returned %+v, want %+v", clients, want)
	}
}

func TestRepositoriesService_ListDependents(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoDependent{{Repo: &repo.Repository{URI: "r2"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryDependents, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	dependents, _, err := client.Repositories.ListDependents(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListDependents returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(dependents, want) {
		t.Errorf("Repositories.ListDependents returned %+v, want %+v", dependents, want)
	}
}

func TestRepositoriesService_ListDependencies(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoDependency{{Repo: &repo.Repository{URI: "r2"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryDependencies, map[string]string{"RepoSpec": "r.com/x", "Rev": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	dependencies, _, err := client.Repositories.ListDependencies(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListDependencies returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(dependencies, want) {
		t.Errorf("Repositories.ListDependencies returned %+v, want %+v", dependencies, want)
	}
}

func TestRepositoriesService_ListByContributor(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoContribution{{Repo: &repo.Repository{URI: "r.com/x"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.PersonRepositoryContributions, map[string]string{"PersonSpec": "a"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"NoFork": "true"})

		writeJSON(w, want)
	})

	repos, _, err := client.Repositories.ListByContributor(PersonSpec{Login: "a"}, &RepositoryListByContributorOptions{NoFork: true})
	if err != nil {
		t.Errorf("Repositories.ListByContributor returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.ListByContributor returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_ListByClient(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoUsageByClient{{DefRepo: &repo.Repository{URI: "r.com/x"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.PersonRepositoryDependencies, map[string]string{"PersonSpec": "a"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	repos, _, err := client.Repositories.ListByClient(PersonSpec{Login: "a"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListByClient returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.ListByClient returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_ListByRefdAuthor(t *testing.T) {
	setup()
	defer teardown()

	want := []*AugmentedRepoUsageOfAuthor{{Repo: &repo.Repository{URI: "r.com/x"}}}

	var called bool
	mux.HandleFunc(urlPath(t, router.PersonRepositoryDependents, map[string]string{"PersonSpec": "a"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	repos, _, err := client.Repositories.ListByRefdAuthor(PersonSpec{Login: "a"}, nil)
	if err != nil {
		t.Errorf("Repositories.ListByRefdAuthor returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.ListByRefdAuthor returned %+v, want %+v", repos, want)
	}
}

func normTime(c *Commit) {
	c.Author.Date = c.Author.Date.In(time.UTC)
	if c.Committer != nil {
		c.Committer.Date = c.Committer.Date.In(time.UTC)
	}
}
