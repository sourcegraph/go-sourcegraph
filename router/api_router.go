package router

import (
	"net/url"

	"github.com/sqs/mux"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

const (
	Build        = "build"
	BuildLog     = "build.log"
	Builds       = "builds"
	BuildTasks   = "build.tasks"
	BuildTaskLog = "build.task.log"

	Org               = "org"
	OrgMembers        = "org.members"
	OrgSettings       = "org.settings"
	OrgSettingsUpdate = "org.settings.update"

	People                        = "people"
	Person                        = "person"
	PersonOrgs                    = "person.orgs"
	PersonAuthors                 = "person.authors"
	PersonClients                 = "person.clients"
	PersonEmails                  = "person.emails"
	PersonFromGitHub              = "person.from-github"
	PersonRepositoryContributions = "person.repo-contributions"
	PersonRepositoryDependencies  = "person.repo-dependencies"
	PersonRepositoryDependents    = "person.repo-dependents"
	PersonRefreshProfile          = "person.refresh-profile"
	PersonSettings                = "person.settings"
	PersonSettingsUpdate          = "person.settings.update"
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

	RepositorySettings       = "repo.settings"
	RepositorySettingsUpdate = "repo.settings.update"

	Unit  = "unit"
	Units = "units"

	Search = "search"

	Snippet = "snippet"

	Defs          = "defs"
	Def           = "def"
	DefBySID      = "def-by-sid"
	DefExamples   = "def.examples"
	DefAuthors    = "def.authors"
	DefClients    = "def.clients"
	DefDependents = "def.dependents"

	ExtGitHubReceiveWebhook = "ext.github.receive-webhook"

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
	repo.Path("/.settings").Methods("GET").Name(RepositorySettings)
	repo.Path("/.settings").Methods("PUT").Name(RepositorySettingsUpdate)

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
	person.Path("/orgs").Methods("GET").Name(PersonOrgs)
	person.Path("/clients").Methods("GET").Name(PersonClients)
	person.Path("/authors").Methods("GET").Name(PersonAuthors)
	person.Path("/emails").Methods("GET").Name(PersonEmails)
	person.Path("/repo-contributions").Methods("GET").Name(PersonRepositoryContributions)
	person.Path("/repo-dependencies").Methods("GET").Name(PersonRepositoryDependencies)
	person.Path("/repo-dependents").Methods("GET").Name(PersonRepositoryDependents)
	person.Path("/external-profile").Methods("PUT").Name(PersonRefreshProfile)
	person.Path("/stats").Methods("PUT").Name(PersonComputeStats)
	person.Path("/settings").Methods("GET").Name(PersonSettings)
	person.Path("/settings").Methods("PUT").Name(PersonSettingsUpdate)
	m.Path("/external-users/github/{GitHubUserSpec}").Methods("GET").Name(PersonFromGitHub)

	orgPath := "/orgs/{OrgSpec}"
	m.Path(orgPath).Methods("GET").Name(Org)
	org := m.PathPrefix(orgPath).Subrouter()
	org.Path("/settings").Methods("GET").Name(OrgSettings)
	org.Path("/settings").Methods("PUT").Name(OrgSettingsUpdate)
	org.Path("/members").Methods("GET").Name(OrgMembers)

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

	m.Path("/.units").Methods("GET").Name(Units)
	unitPath := `/.units/.{UnitType}/{Unit:.*}`
	repo.Path(unitPath).Methods("GET").Name(Unit)

	m.Path("/ext/github/webhook").Methods("POST").Name(ExtGitHubReceiveWebhook)

	return m
}

func URIToDef(key graph.DefKey) *url.URL {
	return URITo(Def, "Repo", string(key.Repo), "UnitType", key.UnitType, "Unit", key.Unit, "Path", string(key.Path))
}

func URITo(routeName string, params ...string) *url.URL {
	return URLTo(APIRouter, routeName, params...)
}
