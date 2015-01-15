package sourcegraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
)

// A Token is the smallest indivisible component of a query, either a
// term or a "field:val" specifier (e.g., "repo:example.com/myrepo").
type Token interface {
	// String returns the string representation of the term.
	String() string
}

// A Term is a query term token. It is either a word or an arbitrary
// string (if quoted in the raw query).
type Term string

func (t Term) String() string {
	if strings.Contains(string(t), " ") {
		return `"` + string(t) + `"`
	}
	return string(t)
}

func (t Term) UnquotedString() string { return string(t) }

// An AnyToken is a token that has not yet been resolved into another
// token type. It resolves to Term if it can't be resolved to another
// token type.
type AnyToken string

func (u AnyToken) String() string { return string(u) }

// A RepoToken represents a repository, although it does not
// necessarily uniquely identify the repository. It consists of any
// number of slash-separated path components, such as "a/b" or
// "github.com/foo/bar".
type RepoToken struct {
	URI string

	Repo *Repo `json:",omitempty"`
}

func (t RepoToken) String() string { return t.URI }

func (t RepoToken) Spec() RepoSpec {
	var rid int
	if t.Repo != nil {
		rid = t.Repo.RID
	}
	return RepoSpec{URI: t.URI, RID: rid}
}

// A RevToken represents a specific revision (either a revspec or a
// commit ID) of a repository (which must be specified by a previous
// RepoToken in the query).
type RevToken struct {
	Rev string // Rev is either a revspec or commit ID

	Commit *Commit `json:",omitempty"`
}

func (t RevToken) String() string { return ":" + t.Rev }

// A UnitToken represents a source unit in a repository.
type UnitToken struct {
	// Type is the type of the source unit (e.g., GoPackage).
	Type string

	// Name is the name of the source unit (e.g., mypkg).
	Name string

	// Unit is the source unit object.
	Unit *unit.RepoSourceUnit
}

func (t UnitToken) String() string { return "~" + t.Name + "@" + t.Type }

type FileToken struct {
	Path string

	Entry *vcsclient.TreeEntry
}

func (t FileToken) String() string { return "/" + filepath.Clean(t.Path) }

// A UserToken represents a user or org, although it does not
// necessarily uniquely identify one. It consists of the string "@"
// followed by a full or partial user/org login.
type UserToken struct {
	Login string

	User *User `json:",omitempty"`
}

func (t UserToken) String() string { return "@" + t.Login }

// Tokens wraps a list of tokens and adds some helper methods. It also
// serializes to JSON with "Type" fields added to each token and
// deserializes that same JSON back into a typed list of tokens.
type Tokens []Token

func (d Tokens) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(([]Token)(d))
	if err != nil {
		return nil, err
	}
	var toks []interface{}
	if err := json.Unmarshal(b, &toks); err != nil {
		return nil, err
	}
	for i, tok := range toks {
		ttype := TokenType(d[i])
		switch tok := tok.(type) {
		case string:
			toks[i] = map[string]string{"Type": ttype, "String": tok}
		case map[string]interface{}:
			tok["Type"] = ttype
		}
	}
	return json.Marshal(toks)
}

func (d *Tokens) UnmarshalJSON(b []byte) error {
	var jtoks []jsonToken
	if err := json.Unmarshal(b, &jtoks); err != nil {
		return err
	}
	if jtoks == nil {
		*d = nil
	} else {
		*d = make(Tokens, len(jtoks))
		for i, jtok := range jtoks {
			(*d)[i] = jtok.Token
		}
	}
	return nil
}

func (d Tokens) RawQueryString() string { return Join(d).String }

type jsonToken struct {
	Token `json:",omitempty"`
}

func (t jsonToken) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(t.Token)
	if err != nil {
		return nil, err
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	if t.Token != nil {
		tokType := TokenType(t.Token)
		switch vv := v.(type) {
		case string:
			v = map[string]string{"Type": tokType, "String": vv}
		case map[string]interface{}:
			vv["Type"] = tokType
		}
	}
	return json.Marshal(v)
}

func (t *jsonToken) UnmarshalJSON(b []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	tok, err := toTypedToken(v)
	if err != nil {
		return err
	}
	*t = jsonToken{tok}
	return nil
}

func toTypedToken(tokJSON map[string]interface{}) (Token, error) {
	if tokJSON == nil {
		return nil, nil
	}
	typ, ok := tokJSON["Type"].(string)
	if !ok {
		return nil, errors.New("unmarshal Tokens: no 'Type' field in token")
	}
	delete(tokJSON, "Type")

	var tok interface{}
	switch typ {
	case "Term", "AnyToken":
		s, _ := tokJSON["String"].(string)
		switch typ {
		case "Term":
			tok = Term(s)
		case "AnyToken":
			tok = AnyToken(s)
		}
		return tok.(Token), nil

	case "RepoToken":
		tok = &RepoToken{}
	case "RevToken":
		tok = &RevToken{}
	case "UnitToken":
		tok = &UnitToken{}
	case "FileToken":
		tok = &FileToken{}
	case "UserToken":
		tok = &UserToken{}
	default:
		return nil, fmt.Errorf("unmarshal Tokens: unrecognized Type %q", typ)
	}
	tmpJSON, err := json.Marshal(tokJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tmpJSON, tok); err != nil {
		return nil, err
	}
	tok = reflect.ValueOf(tok).Elem().Interface() // deref
	return tok.(Token), nil
}

func TokenType(tok Token) string {
	return strings.Replace(strings.Replace(reflect.ValueOf(tok).Type().String(), "*", "", -1), "sourcegraph.", "", -1)
}
