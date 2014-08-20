package router

import (
	"net/url"

	"github.com/sqs/mux"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

const (
	AssistGoto = "assist.goto"
	AssistInfo = "assist.info"

	Build        = "build"
	BuildLog     = "build.log"
	Builds       = "builds"
	BuildTasks   = "build.tasks"
	BuildTaskLog = "build.task.log"

	People                        = "people"
	Person                        = "person"
	PersonAuthors                 = "person.authors"
	PersonClients                 = "person.clients"
	PersonFromGitHub              = "person.from-github"
	PersonOwnedRepositories       = "person.owned-repositories"
	PersonRepositoryContributions = "person.repo-contributions"
	PersonRepositoryDependencies  = "person.repo-dependencies"
	PersonRepositoryDependents    = "person.repo-dependents"
	PersonRefreshProfile          = "person.refresh-profile"
	PersonComputeStats            = "person.compute-stats"

	Repositories             = "repos"
	RepositoriesCreate       = "repos.create"
	RepositoriesGetOrCreate  = "repos.get-or-create"
	Repository               = "repo"
	RepositoryAuthors        = "repo.authors"
	RepositoryClients        = "repo.clients"
	RepositoryDependents     = "repo.dependents"
	RepositoryDependencies   = "repo.dependencies"
	RepositoryBadge          = "repo.badge"
	RepositoryBadges         = "repo.badges"
	RepositoryCounter        = "repo.counter"
	RepositoryCounters       = "repo.counters"
	RepositoryReadme         = "repo.readme"
	RepositoryBuilds         = "repo.builds"
	RepositoryBuildsCreate   = "repo.builds.create"
	RepositoryBuildDataEntry = "repo.build-data.entry"
	RepositoryDocPage        = "repo.doc-page"
	RepositoryTreeEntry      = "repo.tree.entry"
	RepositoryRefreshProfile = "repo.refresh-profile"
	RepositoryRefreshVCSData = "repo.refresh-vcs-data"
	RepositoryComputeStats   = "repo.compute-stats"

	Unit  = "unit"
	Units = "units"

	Search = "search"

	Snippet = "snippet"

	Defs               = "defs"
	Def                = "def"
	DefBySID           = "def-by-sid"
	DefExamples        = "def.examples"
	DefAuthors         = "def.authors"
	DefClients         = "def.clients"
	DefDependents      = "def.dependents"
	DefImplementations = "def.implementations"
	DefInterfaces      = "def.interfaces"

	// Redirects for old routes.
	RedirectOldRepositoryBadgesAndCounters = "repo.redirect-old-badges-and-counters"
)

var APIRouter = NewAPIRouter("/api")

// NewAPIRouter creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
//
// It is in a separate package from app so that other packages may use it to
// generate URLs without resulting in Go import cycles (and so we can release
// the router as open-source to support our client library).
func NewAPIRouter(pathPrefix string) *mux.Router {
	m := mux.NewRouter()

	if pathPrefix != "" && pathPrefix != "/" {
		m = m.PathPrefix(pathPrefix).Subrouter()
	}

	m.StrictSlash(true)

	m.Path("/assist/goto").Methods("GET").Name(AssistGoto)
	m.Path("/assist/info").Methods("GET").Name(AssistInfo)

	m.Path("/builds").Methods("GET").Name(Builds)
	builds := m.PathPrefix("/builds").Subrouter()
	buildPath := "/{BID}"
	builds.Path(buildPath).Methods("GET").Name(Build)
	build := builds.PathPrefix(buildPath).Subrouter()
	build.Path("/log").Methods("GET").Name(BuildLog)
	build.Path("/tasks").Methods("GET").Name(BuildTasks)
	build.Path("/tasks/{TaskID}/log").Methods("GET").Name(BuildTaskLog)

	m.Path("/repos").Methods("GET").Name(Repositories)
	m.Path("/repos").Methods("POST").Name(RepositoriesCreate)

	m.Path("/repos/github.com/{owner:[^/]+}/{repo:[^/]+}/{what:(?:badges|counters)}/{which}.png").Methods("GET").Name(RedirectOldRepositoryBadgesAndCounters)

	// Recognize RepoURIs with 1 or more path components, none of which begins
	// with a ".".
	repoPath := `/repos/` + RepoPathPattern
	m.Path(repoPath).Methods("GET").PostMatchFunc(FixRepoVars).BuildVarsFunc(PrepareRepoRouteVars).Name(Repository)
	m.Path(repoPath).Methods("PUT").PostMatchFunc(FixRepoVars).BuildVarsFunc(PrepareRepoRouteVars).Name(RepositoriesGetOrCreate)
	repo := m.PathPrefix(repoPath).PostMatchFunc(FixRepoVars).BuildVarsFunc(PrepareRepoRouteVars).Subrouter()
	repo.Path("/.authors").Methods("GET").Name(RepositoryAuthors)
	repo.Path("/.clients").Methods("GET").Name(RepositoryClients)
	repo.Path("/.readme").Methods("GET").Name(RepositoryReadme)
	repo.Path("/.dependents").Methods("GET").Name(RepositoryDependents)
	repo.Path("/.dependencies").Methods("GET").Name(RepositoryDependencies)
	repo.Path("/.external-profile").Methods("PUT").Name(RepositoryRefreshProfile)
	repo.Path("/.vcs-data").Methods("PUT").Name(RepositoryRefreshVCSData)
	repo.Path("/.stats").Methods("PUT").Name(RepositoryComputeStats)

	// TODO(new-arch): set up redirects from /badges
	repo.Path("/.badges").Methods("GET").Name(RepositoryBadges)
	repo.Path("/.badges/{Badge}.png").Methods("GET").Name(RepositoryBadge)

	// TODO(new-arch): set up redirects from /counters
	repo.Path("/.counters").Methods("GET").Name(RepositoryCounters)
	repo.Path("/.counters/{Counter}.png").Methods("GET").Name(RepositoryCounter)

	repo.Path("/.builds").Methods("GET").Name(RepositoryBuilds)
	repo.Path("/.builds").Methods("POST").Name(RepositoryBuildsCreate)

	repo.PathPrefix("/.build-data"+TreeEntryPathPattern).PostMatchFunc(FixTreeEntryVars).BuildVarsFunc(PrepareTreeEntryRouteVars).Methods("GET", "PUT").Name(RepositoryBuildDataEntry)

	repo.Path("/.docs/{Path:.*}").Methods("GET").Name(RepositoryDocPage)

	// See router_util/tree_route.go for an explanation of how we match tree
	// entry routes.
	repo.Path("/.tree" + TreeEntryPathPattern).PostMatchFunc(FixTreeEntryVars).BuildVarsFunc(PrepareTreeEntryRouteVars).Methods("GET").Name(RepositoryTreeEntry)

	m.Path("/people").Methods("GET").Name(People)
	personPath := `/people/` + PersonSpecPattern
	m.Path(personPath).Methods("GET").Name(Person)
	person := m.PathPrefix(personPath).Subrouter()
	person.Path("/clients").Methods("GET").Name(PersonClients)
	person.Path("/authors").Methods("GET").Name(PersonAuthors)
	person.Path("/repositories").Methods("GET").Name(PersonOwnedRepositories)
	person.Path("/repo-contributions").Methods("GET").Name(PersonRepositoryContributions)
	person.Path("/repo-dependencies").Methods("GET").Name(PersonRepositoryDependencies)
	person.Path("/repo-dependents").Methods("GET").Name(PersonRepositoryDependents)
	person.Path("/external-profile").Methods("PUT").Name(PersonRefreshProfile)
	person.Path("/stats").Methods("PUT").Name(PersonComputeStats)
	m.Path("/external-users/github/{GitHubUserSpec}").Methods("GET").Name(PersonFromGitHub)

	m.Path("/search").Methods("GET").Name(Search)

	m.Path("/snippet").Methods("GET", "POST", "ORIGIN").Name(Snippet)

	m.Path("/.defs").Methods("GET").Name(Defs)
	m.Path(`/.defs/{SID:\d+}`).Methods("GET").Name(DefBySID)

	// See router_util/def_route.go for an explanation of how we match def
	// routes.
	defPath := `/.defs/` + DefPathPattern
	repo.Path(defPath).Methods("GET").PostMatchFunc(FixDefUnitVars).BuildVarsFunc(PrepareDefRouteVars).Name(Def)
	def := repo.PathPrefix(defPath).PostMatchFunc(FixDefUnitVars).BuildVarsFunc(PrepareDefRouteVars).Subrouter()
	def.Path("/.examples").Methods("GET").Name(DefExamples)
	def.Path("/.authors").Methods("GET").Name(DefAuthors)
	def.Path("/.clients").Methods("GET").Name(DefClients)
	def.Path("/.dependents").Methods("GET").Name(DefDependents)
	def.Path("/.implementations").Methods("GET").Name(DefImplementations)
	def.Path("/.interfaces").Methods("GET").Name(DefInterfaces)

	m.Path("/.units").Methods("GET").Name(Units)
	unitPath := `/.units/.{UnitType}/{Unit:.*}`
	repo.Path(unitPath).Methods("GET").Name(Unit)

	return m
}

func URIToDef(key graph.DefKey) *url.URL {
	return URITo(Def, "Repo", string(key.Repo), "UnitType", key.UnitType, "Unit", key.Unit, "Path", string(key.Path))
}

func URITo(routeName string, params ...string) *url.URL {
	return URLTo(APIRouter, routeName, params...)
}
