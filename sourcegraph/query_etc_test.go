package sourcegraph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResolveError_JSON(t *testing.T) {
	rerr := ResolveErrors{
		ResolveError{Token: Term("t"), Reason: "a"},
		ResolveError{Token: RepoToken{URI: "r"}, Reason: "b"},
	}

	rerrJSON, err := json.MarshalIndent(rerr, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	want := `[
  {
    "Token": {
      "String": "t",
      "Type": "Term"
    },
    "Reason": "a"
  },
  {
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
