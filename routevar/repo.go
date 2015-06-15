package routevar

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"

	"github.com/sourcegraph/mux"
)

var (
	// Repo captures RepoSpec strings in URL routes.
	Repo = `{Repo:` + namedToNonCapturingGroups(spec.RepoPattern) + `}`

	// RepoRev captures RepoRevSpec strings in URL routes.
	RepoRev = Repo + `{ResolvedRev:(?:@` + namedToNonCapturingGroups(spec.ResolvedRevPattern) + `)?}`
)

// FixRepoRevVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a RepoRev.
func FixRepoRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	var keep bool
	if rrev, present := match.Vars["ResolvedRev"]; present {
		rrev = strings.TrimPrefix(rrev, "@")
		if rrev != "" {
			rev, commitID, err := spec.ParseResolvedRev(rrev)
			if err == nil {
				if rev != "" {
					match.Vars["Rev"] = rev
				}
				if commitID != "" {
					match.Vars["CommitID"] = commitID
				}
			} else {
				keep = true
			}
		}
	}

	if !keep {
		delete(match.Vars, "ResolvedRev")
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
