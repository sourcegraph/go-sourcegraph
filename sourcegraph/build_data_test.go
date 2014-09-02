package sourcegraph

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

func TestBuildDataService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := []byte("hello")

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryBuildDataEntry, map[string]string{"RepoSpec": "r.com/x", "Rev": "c", "Path": "a/b"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		w.Write(want)
	})

	file, _, err := client.BuildData.Get(BuildDataFileSpec{RepoRev: RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, Path: "a/b"})
	if err != nil {
		t.Errorf("BuildData.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(file, want) {
		t.Errorf("BuildData.Get returned %+v, want %+v", file, want)
	}
}

func TestBuildDataService_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*buildstore.BuildDataFileInfo{{Path: "a/b", CommitID: "c"}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryBuildDataEntry, map[string]string{"RepoSpec": "r.com/x", "Rev": "c", "Path": "."}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	files, _, err := client.BuildData.List(RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, nil)
	if err != nil {
		t.Errorf("BuildData.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normalizeBuildDataTime(files...)
	normalizeBuildDataTime(want...)
	if !reflect.DeepEqual(files, want) {
		t.Errorf("BuildData.List returned %+v, want %+v", files, want)
	}
}

func TestBuildDataService_Upload(t *testing.T) {
	setup()
	defer teardown()

	want := []byte("hello")

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryBuildDataEntry, map[string]string{"RepoSpec": "r.com/x", "Rev": "c", "Path": "a/b"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "PUT")
	})

	_, err := client.BuildData.Upload(BuildDataFileSpec{RepoRev: RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "c"}, Path: "a/b"}, ioutil.NopCloser(bytes.NewReader(want)))
	if err != nil {
		t.Errorf("BuildData.Upload returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func normalizeBuildDataTime(bs ...*buildstore.BuildDataFileInfo) {
	for _, b := range bs {
		normalizeTime(&b.ModTime)
	}
}
