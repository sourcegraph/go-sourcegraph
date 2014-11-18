package router

import (
	"net/http"
	"strings"

	"github.com/sqs/mux"
)

// RepoSpecPattern is the path pattern for encoding RepoSpec.
//
// TODO(sqs): match the "R$rid" format too.
var RepoSpecPathPattern = `{RepoSpec:(?:(?:[^/.@][^/@]*/)+(?:[^/.@][^/@]*))|(?:R\$\d+)}`

// RepoRevSpecPattern is the path pattern for encoding RepoRevSpec.
var RepoRevSpecPattern = RepoSpecPathPattern + `{Rev:(?:@(?:(?:[^/@]*(?:/[^.@/]+)*)))?}`

// FixRepoRevSpecVars is a mux.PostMatchFunc that cleans and normalizes the
// RepoRevSpecPattern vars.
func FixRepoRevSpecVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	if rev, present := match.Vars["Rev"]; present {
		if rev == "" {
			delete(match.Vars, "Rev")
		} else {
			match.Vars["Rev"] = strings.TrimPrefix(rev, "@")
		}
	}
}

// PrepareRepoRevSpecRouteVars is a mux.BuildVarsFunc that converts
// from a RepoRevSpec's route vars to components used to generate
// routes.
func PrepareRepoRevSpecRouteVars(vars map[string]string) map[string]string {
	if rev, present := vars["Rev"]; !present {
		vars["Rev"] = ""
	} else if rev != "" {
		vars["Rev"] = "@" + rev
	}
	return vars
}
