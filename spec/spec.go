package spec

import "fmt"

const (
	hostComponent = `[a-zA-Z-]+`
	host          = `(?:(?:` + hostComponent + `\.)*` + hostComponent + `)`
)

// InvalidError occurs when a spec string is invalid.
type InvalidError struct {
	Type  string // RepoSpec, UserSpec, etc.
	Input string // the original string input
	Err   error  // underlying error (nil for routine regexp match failures)
}

func (e InvalidError) Error() string {
	str := fmt.Sprintf("invalid input for %s: %q", e.Type, e.Input)
	if e.Err != nil {
		str += " " + e.Err.Error()
	}
	return str
}
