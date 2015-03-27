package sourcegraph

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

func TestNoteSpec_RouteVars(t *testing.T) {
	tests := []struct {
		spec NoteSpec
		want map[string]string
	}{
		{NoteSpec{ID: 1}, map[string]string{"NoteSpec": "1"}},
	}
	for _, test := range tests {
		routeVars := test.spec.RouteVars()
		if !reflect.DeepEqual(routeVars, test.want) {
			t.Errorf("%+v: got %v, want %v", test.spec, routeVars, test.want)
		}
	}
}

func TestUnmarshalNoteSpec(t *testing.T) {
	tests := []struct {
		routeVars map[string]string
		want      *NoteSpec
	}{
		{map[string]string{"NoteSpec": "1"}, &NoteSpec{ID: 1}},
		{map[string]string{"NoteSpec": "a"}, nil},
		{map[string]string{"foo": "bar"}, nil},
	}
	for _, test := range tests {
		noteSpec := UnmarshalNoteSpec(test.routeVars)
		if !reflect.DeepEqual(noteSpec, test.want) {
			t.Errorf("%v: got %+v, want %+v", test.routeVars, noteSpec, test.want)
		}
	}
}

func TestNotesService_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*Note{&Note{ID: 1}}

	var called bool
	mux.HandleFunc(urlPath(t, router.Notes, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"Sort":      "name",
			"Direction": "asc",
			"PerPage":   "1",
			"Page":      "2",
		})

		writeJSON(w, want)
	})

	notes, _, err := client.Notes.List(&NotesListOptions{
		Sort:        "name",
		Direction:   "asc",
		ListOptions: ListOptions{PerPage: 1, Page: 2},
	})
	if err != nil {
		t.Errorf("Notes.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normNote(want...)
	if !reflect.DeepEqual(notes, want) {
		t.Errorf("Notes.List: got %+v, want %+v", notes, want)
	}
}

func TestNotesService_Create(t *testing.T) {
	setup()
	defer teardown()

	want := &Note{ID: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.NotesCreate, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")
		wantBody, _ := json.Marshal(want)
		testBody(t, r, string(wantBody)+"\n")

		writeJSON(w, want)
	})

	note, _, err := client.Notes.Create(&Note{ID: 1})
	if err != nil {
		t.Errorf("Notes.Create returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normNote(want)
	if !reflect.DeepEqual(note, want) {
		t.Errorf("Notes.Create returned %+v, want %+v", note, want)
	}
}

func normNote(n ...*Note) {
	for _, n := range n {
		n.CreatedAt = n.CreatedAt.UTC()
		if n.UpdatedAt != nil {
			tmp := n.UpdatedAt.UTC()
			n.UpdatedAt = &tmp
		}
		if n.ClosedAt != nil {
			tmp := n.ClosedAt.UTC()
			n.ClosedAt = &tmp
		}
	}
}
