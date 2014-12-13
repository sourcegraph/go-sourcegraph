package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func TestUnitsService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &unit.RepoSourceUnit{
		Repo:     "x.com/r",
		CommitID: "c",
		UnitType: "t",
		Unit:     "u",
		Data:     []byte(`{"k":"v"}`),
	}

	var called bool
	mux.HandleFunc(urlPath(t, router.Unit, map[string]string{"RepoSpec": "x.com/r", "Rev": "c", "UnitType": "t", "Unit": "u"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	unit, _, err := client.Units.Get(UnitSpec{RepoRevSpec: RepoRevSpec{RepoSpec: RepoSpec{URI: "x.com/r"}, Rev: "c"}, UnitType: "t", Unit: "u"})
	if err != nil {
		t.Errorf("Unit.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(unit, want) {
		t.Errorf("Units.Get returned %+v, want %+v", unit, want)
	}
}

func TestUnitsService_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*unit.RepoSourceUnit{
		{
			Repo:     "x.com/r",
			UnitType: "t",
			Data:     []byte(`{}`),
		},
	}

	var called bool
	mux.HandleFunc(urlPath(t, router.Units, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"RepositoryURI": "r1",
			"PerPage":       "1",
			"Page":          "2",
		})

		writeJSON(w, want)
	})

	units, _, err := client.Units.List(&UnitListOptions{
		RepositoryURI: "r1",
		ListOptions:   ListOptions{PerPage: 1, Page: 2},
	})
	if err != nil {
		t.Errorf("Units.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(units, want) {
		t.Errorf("Units.List returned %+v, want %+v", units, want)
	}
}
