package router

import (
	"net/http"
	"strings"

	"github.com/sqs/mux"
)

// SymbolPathPattern is the path pattern for symbols.
//
// We want the symbol routes to match the 2 following forms:
//
//   1. /.MyUnitType/.def/MySymbol (i.e., Unit == ".")
//   2. /.MyUnitType/MyUnitPath1/.def/MySymbol (i.e., Unit == "MyUnitPath1")
//
// To achieve this, we use a non-picky regexp for rawUnit and then sort it
// out in the FixSymbolUnitVars PostMatchFunc.
var SymbolPathPattern = `.{UnitType}/{rawUnit:.*}.def{Path:(?:(?:/(?:[^/.][^/]*/)*(?:[^/.][^/]*))|)}`

// FixSymbolUnitVars is a mux.PostMatchFunc that cleans up the dummy rawUnit route
// variable matched by SymbolPathPattern. See the docs for SymbolPathPattern for
// more information.
func FixSymbolUnitVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	match.Vars["Path"] = strings.TrimPrefix(match.Vars["Path"], "/")
	if path := match.Vars["Path"]; path == "" {
		match.Vars["Path"] = "."
	}
	match.Vars["Path"] = pathUnescape(match.Vars["Path"])

	rawUnit := match.Vars["rawUnit"]
	if rawUnit == "" {
		match.Vars["Unit"] = "."
	} else {
		match.Vars["Unit"] = strings.TrimSuffix(rawUnit, "/")
	}
	delete(match.Vars, "rawUnit")
}

// PrepareSymbolRouteVars is a mux.BuildVarsFunc that converts from a "Unit"
// route variable to the dummy "rawUnit" route variable that actually appears in
// the route regexp pattern.
func PrepareSymbolRouteVars(vars map[string]string) map[string]string {
	if path := vars["Path"]; path == "." {
		vars["Path"] = ""
	} else if path != "" {
		vars["Path"] = "/" + path
	}

	vars["Path"] = pathEscape(vars["Path"])

	if unit := vars["Unit"]; unit == "." {
		vars["rawUnit"] = ""
	} else {
		vars["rawUnit"] = unit + "/"
	}
	delete(vars, "Unit")

	return vars
}

// pathEscape is a limited version of url.QueryEscape that only escapes '?'.
func pathEscape(p string) string {
	return strings.Replace(p, "?", "%3F", -1)
}

// pathUnescape is a limited version of url.QueryEscape that only unescapes '?'.
func pathUnescape(p string) string {
	return strings.Replace(p, "%3F", "?", -1)
}
