package routevar

import "sourcegraph.com/sourcegraph/go-sourcegraph/spec"

var (
	// User captures UserSpec strings in URL routes.
	User = `{User:` + namedToNonCapturingGroups(spec.UserPattern) + `}`

	// Person captures PersonSpec strings in URL routes.
	Person = `{Person:` + namedToNonCapturingGroups(spec.UserPattern) + `}`
)
