package sourcegraph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTokens_JSON(t *testing.T) {
	tokens := Tokens{
		AnyToken("a"),
		Term("b"),
		Term(""),
		RepoToken{URI: "r"},
		RevToken{Rev: "v"},
		FileToken{Path: "p"},
		UserToken{Login: "u"},
	}

	b, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	wantJSON := `[
  {
    "String": "a",
    "Type": "AnyToken"
  },
  {
    "String": "b",
    "Type": "Term"
  },
  {
    "String": "",
    "Type": "Term"
  },
  {
    "Type": "RepoToken",
    "URI": "r"
  },
  {
    "Rev": "v",
    "Type": "RevToken"
  },
  {
    "Entry": null,
    "Path": "p",
    "Type": "FileToken"
  },
  {
    "Login": "u",
    "Type": "UserToken"
  }
]`
	if string(b) != wantJSON {
		t.Errorf("got JSON\n%s\n\nwant JSON\n%s", b, wantJSON)
	}

	var tokens2 Tokens
	if err := json.Unmarshal(b, &tokens2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tokens2, tokens) {
		t.Errorf("got tokens\n%+v\n\nwant tokens\n%+v", tokens2, tokens)
	}
}

func TestTokens_nil(t *testing.T) {
	tokens := Tokens(nil)

	b, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	wantJSON := `null`
	if string(b) != wantJSON {
		t.Errorf("got JSON\n%s\n\nwant JSON\n%s", b, wantJSON)
	}
}
