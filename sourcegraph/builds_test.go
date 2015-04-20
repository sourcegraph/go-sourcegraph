package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/db_common"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func TestBuildsService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &Build{BID: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.Build, map[string]string{"RepoSpec": "r.com/x", "BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	build, _, err := client.Builds.Get(BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, nil)
	if err != nil {
		t.Errorf("Builds.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(build, want)
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Builds.Get returned %+v, want %+v", build, want)
	}
}

func TestBuildsService_GetRepoBuildInfo(t *testing.T) {
	setup()
	defer teardown()

	want := &RepoBuildInfo{
		Exact:                &Build{BID: 1},
		LastSuccessful:       &Build{BID: 2},
		CommitsBehind:        3,
		LastSuccessfulCommit: &vcs.Commit{Message: "m"},
	}
	normalizeBuildTime(want.Exact)
	normalizeBuildTime(want.LastSuccessful)

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBuildInfo, map[string]string{"RepoSpec": "r.com/x", "Rev": "r"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	buildInfo, _, err := client.Builds.GetRepoBuildInfo(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "r"}, nil)
	if err != nil {
		t.Errorf("Builds.GetRepoBuildInfo returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(buildInfo.Exact)
	normalizeBuildTime(buildInfo.LastSuccessful)
	if !reflect.DeepEqual(buildInfo.Exact, want.Exact) {
		t.Errorf("Builds.GetRepoBuildInfo returned %+v, want %+v", buildInfo.Exact, want.Exact)
	}
}

func TestBuildsService_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*Build{{BID: 1}}

	var called bool
	mux.HandleFunc(urlPath(t, router.Builds, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	builds, _, err := client.Builds.List(nil)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(builds...)
	normalizeBuildTime(want...)
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_Create(t *testing.T) {
	setup()
	defer teardown()

	config := &BuildCreateOptions{BuildConfig: BuildConfig{Import: true, Queue: true}, Force: true}
	want := &Build{BID: 1, Repo: "r.com/x"}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBuildsCreate, map[string]string{"RepoSpec": "r.com/x", "Rev": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		testBody(t, r, `{"Import":true,"Queue":true,"UseCache":false,"Priority":0,"Force":true}`+"\n")

		writeJSON(w, want)
	})

	build_, _, err := client.Builds.Create(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, config)
	if err != nil {
		t.Errorf("Builds.Create returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(build_)
	normalizeBuildTime(want)
	if !reflect.DeepEqual(build_, want) {
		t.Errorf("Builds.Create returned %+v, want %+v", build_, want)
	}
}

func TestBuildsService_Update(t *testing.T) {
	setup()
	defer teardown()

	update := BuildUpdate{Host: String("h")}
	want := &Build{BID: 1, Repo: "r.com/x"}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildUpdate, map[string]string{"RepoSpec": "r.com/x", "BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
		testBody(t, r, `{"StartedAt":null,"EndedAt":null,"HeartbeatAt":null,"Host":"h","Success":null,"Purged":null,"Failure":null,"Killed":null,"Priority":null}`+"\n")

		writeJSON(w, want)
	})

	build, _, err := client.Builds.Update(BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, update)
	if err != nil {
		t.Errorf("Builds.Update returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(build)
	normalizeBuildTime(want)
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Builds.Update returned %+v, want %+v", build, want)
	}
}

func TestBuildsService_UpdateTask(t *testing.T) {
	setup()
	defer teardown()

	update := TaskUpdate{Success: Bool(true)}
	want := &BuildTask{BID: 1, TaskID: 456, CreatedAt: db_common.NullTime{}}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildTaskUpdate, map[string]string{"RepoSpec": "r.com/x", "BID": "1", "TaskID": "456"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
		testBody(t, r, `{"StartedAt":null,"EndedAt":null,"Success":true,"Failure":null}`+"\n")

		writeJSON(w, want)
	})

	task, _, err := client.Builds.UpdateTask(TaskSpec{BuildSpec: BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, TaskID: 456}, update)
	if err != nil {
		t.Errorf("Builds.UpdateTask returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
	if !reflect.DeepEqual(task, want) {
		t.Errorf("Builds.UpdateTask returned %+v, want %+v", task, want)
	}
}

func TestBuildsService_CreateTasks(t *testing.T) {
	setup()
	defer teardown()

	create := []*BuildTask{
		{BID: 1, Op: "foo", UnitType: "t", Unit: "u"},
		{BID: 1, Op: "bar", UnitType: "t", Unit: "u"},
	}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildTasksCreate, map[string]string{"RepoSpec": "r.com/x", "BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		testBody(t, r, `[{"Repo":"","BID":1,"UnitType":"t","Unit":"u","Op":"foo","CreatedAt":null,"StartedAt":null,"EndedAt":null,"Queue":false},{"Repo":"","BID":1,"UnitType":"t","Unit":"u","Op":"bar","CreatedAt":null,"StartedAt":null,"EndedAt":null,"Queue":false}]`+"\n")
		writeJSON(w, create)
	})

	tasks, _, err := client.Builds.CreateTasks(BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, create)
	if err != nil {
		t.Errorf("Builds.CreateTasks returned error: %v", err)
	}
	if len(tasks) != len(create) {
		t.Error("len(tasks) != len(create)")
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestBuildsService_GetLog(t *testing.T) {
	setup()
	defer teardown()

	want := &LogEntries{MaxID: "1"}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildLog, map[string]string{"RepoSpec": "r.com/x", "BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	entries, _, err := client.Builds.GetLog(BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, nil)
	if err != nil {
		t.Errorf("Builds.GetLog returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(entries, want) {
		t.Errorf("Builds.GetLog returned %+v, want %+v", entries, want)
	}
}

func TestBuildsService_GetTaskLog(t *testing.T) {
	setup()
	defer teardown()

	want := &LogEntries{MaxID: "1"}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildTaskLog, map[string]string{"RepoSpec": "r.com/x", "BID": "1", "TaskID": "2"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	entries, _, err := client.Builds.GetTaskLog(TaskSpec{BuildSpec: BuildSpec{Repo: RepoSpec{URI: "r.com/x"}, BID: 1}, TaskID: 2}, nil)
	if err != nil {
		t.Errorf("Builds.GetTaskLog returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(entries, want) {
		t.Errorf("Builds.GetTaskLog returned %+v, want %+v", entries, want)
	}
}

func TestBuildsService_DequeueNext(t *testing.T) {
	setup()
	defer teardown()

	want := &Build{BID: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildDequeueNext, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")

		writeJSON(w, want)
	})

	build, _, err := client.Builds.DequeueNext()
	if err != nil {
		t.Errorf("Builds.DequeueNext returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(build, want)
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Builds.DequeueNext returned %+v, want %+v", build, want)
	}
}

func TestBuildsService_DequeueNext_emptyQueue(t *testing.T) {
	setup()
	defer teardown()

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildDequeueNext, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNotFound)
	})

	build, _, err := client.Builds.DequeueNext()
	if err != nil {
		t.Errorf("Builds.DequeueNext returned error: %v", err)
	}
	if build != nil {
		t.Errorf("got build %v, want nil (no builds in queue)", build)
	}

	if !called {
		t.Fatal("!called")
	}

}

func normalizeBuildTime(bs ...*Build) {
	for _, b := range bs {
		if b != nil {
			normalizeTime(&b.CreatedAt)
			normalizeTime(&b.StartedAt.Time)
			normalizeTime(&b.EndedAt.Time)
		}
	}
}
