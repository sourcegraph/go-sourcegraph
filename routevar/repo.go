package routevar

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"

	"github.com/sourcegraph/mux"
)

var (
	// Repo captures RepoSpec strings in URL routes.
	Repo = `{Repo:` + NamedToNonCapturingGroups(spec.RepoPattern) + `}`

	// RepoRev captures RepoRevSpec strings in URL routes.
	RepoRev = Repo + `{ResolvedRev:(?:@` + NamedToNonCapturingGroups(spec.ResolvedRevPattern) + `)?}`
)

// FixRepoRevVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a RepoRev.
func FixRepoRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if rrev, present := match.Vars["ResolvedRev"]; present {
		rrev = strings.TrimPrefix(rrev, "@")
		rev, commitID, err := spec.ParseResolvedRev(rrev)
		if err == nil || rrev == "" {
			// Propagate ResolvedRev if it was set and if parsing
			// failed; otherwise remove it.
			delete(match.Vars, "ResolvedRev")
		}
		if err == nil {
			if rev != "" {
				match.Vars["Rev"] = rev
			}
			if commitID != "" {
				match.Vars["CommitID"] = commitID
			}
		}
	}
}

// PrepareRepoRevRouteVars is a mux.BuildVarsFunc that converts from a
// RepoRevSpec's route vars to components used to generate routes.
func PrepareRepoRevRouteVars(vars map[string]string) map[string]string {
	rrev := spec.ResolvedRevString(vars["Rev"], vars["CommitID"])
	if rrev != "" {
		rrev = "@" + rrev
	}
	vars["ResolvedRev"] = rrev
	return vars
}
