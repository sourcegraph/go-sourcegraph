package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
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
		testFormValues(t, r, values{"Formatted": "true", "StartByte": "0", "EndByte": "0"})

		writeJSON(w, want)
	})

	data, _, err := client.RepositoryTree.Get(TreeEntrySpec{
		RepoRev: RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}, Rev: "v"},
		Path:    "p",
	}, &RepositoryTreeGetOptions{Formatted: true})
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
