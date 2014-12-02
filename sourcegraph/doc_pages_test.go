package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestDocPagesService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &graph.DocPage{Title: "hello"}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoDocPage, map[string]string{"RepoSpec": "r.com/x", "Path": "p"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	docPage, _, err := client.DocPages.Get(DocPageSpec{RepoRev: RepoRevSpec{RepoSpec: RepoSpec{URI: "r.com/x"}}, Path: "p"}, nil)
	if err != nil {
		t.Errorf("DocPages.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(docPage, want) {
		t.Errorf("DocPages.Get returned %+v, want %+v", docPage, want)
	}
}
