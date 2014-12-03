package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

func TestBuildsService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &Build{BID: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.Build, map[string]string{"BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	build, _, err := client.Builds.Get(BuildSpec{BID: 1}, nil)
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

func TestRepoBuildsService_ListByRepo(t *testing.T) {
	setup()
	defer teardown()

	want := []*Build{{BID: 1}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBuilds, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	builds, _, err := client.Builds.ListByRepo(RepoSpec{URI: "r.com/x"}, nil)
	if err != nil {
		t.Errorf("Builds.ListByRepo returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildTime(builds...)
	normalizeBuildTime(want...)
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.ListByRepo returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_Create(t *testing.T) {
	setup()
	defer teardown()

	config := &BuildCreateOptions{BuildConfig: BuildConfig{Import: true, Queue: true, CommitID: "c"}, Force: true}
	want := &Build{BID: 123, Repo: 456}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoBuildsCreate, map[string]string{"RepoSpec": "r.com/x"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		testBody(t, r, `{"Import":true,"Queue":true,"UseCache":false,"Priority":0,"CommitID":"c","Force":true}`+"\n")

		writeJSON(w, want)
	})

	build_, _, err := client.Builds.Create(RepoSpec{URI: "r.com/x"}, config)
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
	want := &Build{BID: 123, Repo: 456}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildUpdate, map[string]string{"BID": "123"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
		testBody(t, r, `{"StartedAt":null,"EndedAt":null,"HeartbeatAt":null,"Host":"h","Success":null,"Purged":null,"Failure":null,"Killed":null,"Priority":null}`+"\n")

		writeJSON(w, want)
	})

	build, _, err := client.Builds.Update(BuildSpec{BID: 123}, update)
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
	want := &BuildTask{BID: 123, TaskID: 456, CreatedAt: time.Time{}.In(time.UTC)}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildTaskUpdate, map[string]string{"BID": "123", "TaskID": "456"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
		testBody(t, r, `{"StartedAt":null,"EndedAt":null,"Success":true,"Failure":null}`+"\n")

		writeJSON(w, want)
	})

	task, _, err := client.Builds.UpdateTask(TaskSpec{BuildSpec: BuildSpec{BID: 123}, TaskID: 456}, update)
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
		{BID: 123, Op: "foo", UnitType: "t", Unit: "u"},
		{BID: 123, Op: "bar", UnitType: "t", Unit: "u"},
	}

	var called bool
	mux.HandleFunc(urlPath(t, router.BuildTasksCreate, map[string]string{"BID": "123"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		testBody(t, r, `[{"BID":123,"UnitType":"t","Unit":"u","Op":"foo","CreatedAt":"0001-01-01T00:00:00Z","StartedAt":null,"EndedAt":null,"Queue":false},{"BID":123,"UnitType":"t","Unit":"u","Op":"bar","CreatedAt":"0001-01-01T00:00:00Z","StartedAt":null,"EndedAt":null,"Queue":false}]`+"\n")
		writeJSON(w, create)
	})

	tasks, _, err := client.Builds.CreateTasks(BuildSpec{BID: 123}, create)
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
	mux.HandleFunc(urlPath(t, router.BuildLog, map[string]string{"BID": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	entries, _, err := client.Builds.GetLog(BuildSpec{BID: 1}, nil)
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
	mux.HandleFunc(urlPath(t, router.BuildTaskLog, map[string]string{"BID": "1", "TaskID": "2"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	entries, _, err := client.Builds.GetTaskLog(TaskSpec{BuildSpec: BuildSpec{BID: 1}, TaskID: 2}, nil)
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

func normalizeBuildTime(bs ...*Build) {
	for _, b := range bs {
		if b != nil {
			normalizeTime(&b.CreatedAt)
			normalizeTime(&b.StartedAt.Time)
			normalizeTime(&b.EndedAt.Time)
		}
	}
}
