package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
)

func TestRepositoryTreeService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &TreeEntry{
		TreeEntry: &vcsclient.TreeEntry{
			Name:     "p",
			Type:     vcsclient.FileEntry,
			Size:     123,
			Contents: []byte("hello"),
		},
	}
	want.ModTime = want.ModTime.In(time.UTC)

	var called bool
	mux.HandleFunc(urlPath(t, router.RepositoryTreeEntry, map[string]string{"RepoSpec": "r.com/x", "Rev": "v", "Path": "p"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"Formatted":          "true",
			"ExpandContextLines": "2",
			"StartByte":          "123",
			"EndByte":            "456",
		})

		writeJSON(w, want)
	})

	opt := &RepositoryTreeGetOptions{
		Formatted: true,
		GetFileOptions: vcsclient.GetFileOptions{
			FileRange:          vcsclient.FileRange{StartByte: 123, EndByte: 456},
			ExpandContextLines: 2,
		},
	}
	data, _, err := client.RepositoryTree.Get(TreeEntrySpec{
		RepoRev: RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "v"},
		Path:    "p",
	}, opt)
	if err != nil {
		t.Errorf("RepositoryTree.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("RepositoryTree.Get returned %+v, want %+v", data, want)
	}
}
