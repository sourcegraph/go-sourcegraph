package sourcegraph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResolveError_JSON(t *testing.T) {
	rerr := ResolveErrors{
		ResolveError{Reason: "a"},
		ResolveError{Token: Term("t"), Reason: "a"},
		ResolveError{Index: 1, Token: Term(""), Reason: "a"},
		ResolveError{Index: 2, Token: RepoToken{URI: "r"}, Reason: "b"},
	}

	rerrJSON, err := json.MarshalIndent(rerr, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	want := `[
  {
    "Token": null,
    "Reason": "a"
  },
  {
    "Token": {
      "String": "t",
      "Type": "Term"
    },
    "Reason": "a"
  },
  {
    "Index": 1,
    "Token": {
      "String": "",
      "Type": "Term"
    },
    "Reason": "a"
  },
  {
    "Index": 2,
    "Token": {
      "Type": "RepoToken",
      "URI": "r"
    },
    "Reason": "b"
  }
]`
	if string(rerrJSON) != want {
		t.Errorf("got JSON\n%s\n\nwant JSON\n%s", rerrJSON, want)
	}

	var rerr2 ResolveErrors
	if err := json.Unmarshal(rerrJSON, &rerr2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rerr2, rerr) {
		t.Errorf("got\n%+v\n\nwant\n%+v", rerr2, rerr)
	}
}
