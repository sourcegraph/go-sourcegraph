package sourcegraph

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

const (
	Git = "git"
	Hg  = "hg"
)

// ErrRenamed is an error type that indicates that a repository was renamed from
// OldURI to NewURI.
type ErrRenamed struct {
	// OldURI is the previous repository URI.
	OldURI string

	// NewURI is the new URI that the repository was renamed to.
	NewURI string
}

func (e ErrRenamed) Error() string {
	return fmt.Sprintf("repository URI %q was renamed to %q; use the new name", e.OldURI, e.NewURI)
}

// ErrNotExist is an error definitively indicating that no such repository
// exists.
var ErrNotExist = errors.New("repository does not exist on external host")

// ErrForbidden is an error indicating that the repository can no longer be
// accessed due to server's refusal to serve it (possibly DMCA takedowns on
// github etc)
var ErrForbidden = errors.New("repository is unavailable")

// ErrNotPersisted is an error indicating that no such repository is persisted
// locally. The repository might exist on a remote host, but it must be
// explicitly added (it will not be implicitly added via a Get call).
var ErrNotPersisted = errors.New("repository is not persisted locally, but it might exist remotely (explicitly add it to check)")

// ErrNotPersisted is an error indicating that repository cannot be created
// without an explicit clone URL, because it has a non-standard URI. It implies
// ErrNotPersisted.
var ErrNonStandardURI = errors.New("cannot infer repository clone URL because repository host is not standard; try adding it explicitly")

// ErrNoRepoBuild indicates that no build could be found for a repo
// revspec.
var ErrNoRepoBuild = errors.New("no suitable repo build found for revspec")

type ErrRedirect struct {
	RedirectURI string
}

func (e ErrRedirect) Error() string {
	return fmt.Sprintf("the repository requested exists at another URI (%s)", e.RedirectURI)
}

var errRedirectMsgPattern = regexp.MustCompile(`the repository requested exists at another URI \(([^\(\)]*)\)`)

func ErrRedirectFromString(msg string) *ErrRedirect {
	if match := errRedirectMsgPattern.FindStringSubmatch(msg); len(match) == 2 {
		return &ErrRedirect{match[1]}
	}
	return nil
}

// IsNotPresent returns whether err is one of ErrNotExist, ErrNotPersisted, or
// ErrRedirected.
func IsNotPresent(err error) bool {
	return err == ErrNotExist || err == ErrNotPersisted
}

func IsForbidden(err error) bool {
	return err == ErrForbidden
}

// ErrNoScheme is an error indicating that a clone URL contained no scheme
// component (e.g., "http://").
var ErrNoScheme = errors.New("clone URL has no scheme")

// ExternalHostTimeout is the timeout for HTTP requests to external repository
// hosts.
var ExternalHostTimeout = time.Second * 7
