package routevar

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"

	"github.com/sourcegraph/mux"
)

var (
	// Repo captures RepoSpec strings in URL routes.
	Repo = `{Repo:` + NamedToNonCapturingGroups(spec.RepoURLPattern) + `}`

	// RepoRev captures RepoRevSpec strings in URL routes.
	RepoRev = Repo + `{ResolvedRev:(?:@` + NamedToNonCapturingGroups(spec.ResolvedRevPattern) + `)?}`
)

// FixRepoVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a Repo.
func FixRepoVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if _, present := match.Vars["Repo"]; present {
		repoSpec := match.Vars["Repo"]
		if !strings.HasPrefix(repoSpec, "src:///") && !strings.HasPrefix(repoSpec, "src://") {
			match.Vars["Repo"] = "src:///" + repoSpec
		}
	}
}

// PrepareRepoRouteVars is a mux.BuildVarsFunc that converts from a
// RepoSpec's route vars to components used to generate routes.
func PrepareRepoRouteVars(vars map[string]string) map[string]string {
	if vars["Repo"] != "" {
		vars["Repo"] = strings.TrimPrefix(vars["Repo"], "src:///")
		vars["Repo"] = strings.TrimPrefix(vars["Repo"], "src://")
	}
	return vars
}

// FixRepoRevVars is a mux.PostMatchFunc that cleans and normalizes
// the route vars pertaining to a RepoRev.
func FixRepoRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	FixRepoVars(req, match, r)
	if _, present := match.Vars["ResolvedRev"]; present {
		match.Vars["ResolvedRev"] = strings.TrimPrefix(match.Vars["ResolvedRev"], "@")
	}
	FixResolvedRevVars(req, match, r)
}

// PrepareRepoRevRouteVars is a mux.BuildVarsFunc that converts from a
// RepoRevSpec's route vars to components used to generate routes.
func PrepareRepoRevRouteVars(vars map[string]string) map[string]string {
	vars = PrepareRepoRouteVars(vars)
	vars = PrepareResolvedRevRouteVars(vars)
	if vars["ResolvedRev"] != "" {
		vars["ResolvedRev"] = "@" + vars["ResolvedRev"]
	}
	return vars
}

// FixResolvedRevVars is a mux.PostMatchFunc that cleans and
// normalizes the route vars pertaining to a ResolvedRev (Rev and CommitID).
func FixResolvedRevVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if rrev, present := match.Vars["ResolvedRev"]; present {
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

// PrepareResolvedRevRouteVars is a mux.BuildVarsFunc that converts
// from a ResolvedRev's component route vars (Rev and CommitID) to a
// single ResolvedRev var.
func PrepareResolvedRevRouteVars(vars map[string]string) map[string]string {
	vars["ResolvedRev"] = spec.ResolvedRevString(vars["Rev"], vars["CommitID"])
	return vars
}
