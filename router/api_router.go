package router

import (
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/routevar"
	"sourcegraph.com/sourcegraph/go-sourcegraph/spec"
)

const (
	Build            = "build"
	BuildDequeueNext = "build.dequeue-next"
	BuildUpdate      = "build.update"
	BuildLog         = "build.log"
	Builds           = "builds"
	BuildTasks       = "build.tasks"
	BuildTaskUpdate  = "build.task"
	BuildTasksCreate = "build.tasks.create"
	BuildTaskLog     = "build.task.log"

	Org               = "org"
	OrgMembers        = "org.members"
	OrgSettings       = "org.settings"
	OrgSettingsUpdate = "org.settings.update"

	Users              = "users"
	User               = "user"
	UserOrgs           = "user.orgs"
	UserEmails         = "user.emails"
	UserSettings       = "user.settings"
	UserSettingsUpdate = "user.settings.update"

	Person = "person"

	Repos              = "repos"
	ReposCreate        = "repos.create"
	Repo               = "repo"
	RepoBadge          = "repo.badge"
	RepoBadges         = "repo.badges"
	RepoCounter        = "repo.counter"
	RepoCounters       = "repo.counters"
	RepoReadme         = "repo.readme"
	RepoBuildsCreate   = "repo.builds.create"
	RepoTreeEntry      = "repo.tree.entry"
	RepoTreeSearch     = "repo.tree.search"
	RepoRefreshVCSData = "repo.refresh-vcs-data"

	RepoSettings       = "repo.settings"
	RepoSettingsUpdate = "repo.settings.update"

	RepoCombinedStatus = "repo.combined-status"
	RepoStatusCreate   = "repo.status.create"

	RepoBuildInfo = "repo.build"

	RepoCommits        = "repo.commits"
	RepoCommit         = "repo.commit"
	RepoCompareCommits = "repo.compare-commits"
	RepoTags           = "repo.tags"
	RepoBranches       = "repo.branches"

	Search         = "search"
	SearchComplete = "search.complete"

	SearchSuggestions = "search.suggestions"

	Defs        = "defs"
	Def         = "def"
	DefRefs     = "def.refs"
	DefExamples = "def.examples"
	DefAuthors  = "def.authors"
	DefClients  = "def.clients"

	Delta                = "delta"
	DeltaUnits           = "delta.units"
	DeltaDefs            = "delta.defs"
	DeltaFiles           = "delta.files"
	DeltaAffectedAuthors = "delta.affected-authors"
	DeltaAffectedClients = "delta.affected-clients"

	Unit  = "unit"
	Units = "units"

	Markdown = "markdown"

	ExtGitHubReceiveWebhook = "ext.github.receive-webhook"

	// Redirects for old routes.
	RedirectOldRepoBadgesAndCounters = "repo.redirect-old-badges-and-counters"
)

// NewAPIRouter creates a new API router with route URL pattern definitions but
// no handlers attached to the routes.
//
// It is in a separate package from app so that other packages may use it to
// generate URLs without resulting in Go import cycles (and so we can release
// the router as open-source to support our client library).
func NewAPIRouter(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	base.StrictSlash(true)

	base.Path("/builds").Methods("GET").Name(Builds)
	builds := base.PathPrefix("/builds").Subrouter()
	builds.Path("/next").Methods("POST").Name(BuildDequeueNext)

	base.Path("/repos").Methods("GET").Name(Repos)
	base.Path("/repos").Methods("POST").Name(ReposCreate)

	base.Path("/repos/github.com/{owner:[^/]+}/{repo:[^/]+}/{what:(?:badges|counters)}/{which}.{Format}").Methods("GET").Name(RedirectOldRepoBadgesAndCounters)

	repoRev := base.PathPrefix(`/repos/` + routevar.RepoRev).PostMatchFunc(routevar.FixRepoRevVars).BuildVarsFunc(routevar.PrepareRepoRevRouteVars).Subrouter()
	repoRev.Path("/.status").Methods("GET").Name(RepoCombinedStatus)
	repoRev.Path("/.status").Methods("POST").Name(RepoStatusCreate)
	repoRev.Path("/.readme").Methods("GET").Name(RepoReadme)
	repoRev.Path("/.badges/{Badge}.{Format}").Methods("GET").Name(RepoBadge)

	// repo contains routes that are NOT specific to a revision. In these routes, the URL may not contain a revspec after the repo (that is, no "github.com/foo/bar@myrevspec").
	repoPath := `/repos/` + routevar.Repo
	base.Path(repoPath).Methods("GET").Name(Repo)
	repo := base.PathPrefix(repoPath).Subrouter()
	repo.Path("/.vcs-data").Methods("PUT").Name(RepoRefreshVCSData)
	repo.Path("/.settings").Methods("GET").Name(RepoSettings)
	repo.Path("/.settings").Methods("PUT").Name(RepoSettingsUpdate)
	repo.Path("/.commits").Methods("GET").Name(RepoCommits)
	repo.Path("/.commits/{Rev:" + spec.PathNoLeadingDotComponentPattern + "}/.compare").Methods("GET").Name(RepoCompareCommits)
	repo.Path("/.commits/{Rev:" + spec.PathNoLeadingDotComponentPattern + "}").Methods("GET").Name(RepoCommit)
	repo.Path("/.branches").Methods("GET").Name(RepoBranches)
	repo.Path("/.tags").Methods("GET").Name(RepoTags)
	repo.Path("/.badges").Methods("GET").Name(RepoBadges)
	repo.Path("/.counters").Methods("GET").Name(RepoCounters)
	repo.Path("/.counters/{Counter}.{Format}").Methods("GET").Name(RepoCounter)

	repoRev.Path("/.build").Methods("GET").Name(RepoBuildInfo)
	repoRev.Path("/.builds").Methods("POST").Name(RepoBuildsCreate)
	buildPath := "/.builds/{CommitID}/{Attempt}"
	repo.Path(buildPath).Methods("GET").Name(Build)
	repo.Path(buildPath).Methods("PUT").Name(BuildUpdate)
	build := repo.PathPrefix(buildPath).Subrouter()
	build.Path("/log").Methods("GET").Name(BuildLog)
	build.Path("/tasks").Methods("GET").Name(BuildTasks)
	build.Path("/tasks").Methods("POST").Name(BuildTasksCreate)
	build.Path("/tasks/{TaskID}").Methods("PUT").Name(BuildTaskUpdate)
	build.Path("/tasks/{TaskID}/log").Methods("GET").Name(BuildTaskLog)

	deltaPath := "/.deltas/{ResolvedRev:" + routevar.NamedToNonCapturingGroups(spec.ResolvedRevPattern) + "}..{DeltaHeadResolvedRev:" + routevar.NamedToNonCapturingGroups(spec.ResolvedRevPattern) + "}"
	repo.Path(deltaPath).Methods("GET").PostMatchFunc(routevar.FixResolvedRevVars).BuildVarsFunc(routevar.PrepareResolvedRevRouteVars).Name(Delta)
	deltas := repo.PathPrefix(deltaPath).PostMatchFunc(routevar.FixResolvedRevVars).BuildVarsFunc(routevar.PrepareResolvedRevRouteVars).Subrouter()
	deltas.Path("/.units").Methods("GET").Name(DeltaUnits)
	deltas.Path("/.defs").Methods("GET").Name(DeltaDefs)
	deltas.Path("/.files").Methods("GET").Name(DeltaFiles)
	deltas.Path("/.affected-authors").Methods("GET").Name(DeltaAffectedAuthors)
	deltas.Path("/.affected-clients").Methods("GET").Name(DeltaAffectedClients)

	// See router_util/tree_route.go for an explanation of how we match tree
	// entry routes.
	repoRev.Path("/.tree" + routevar.TreeEntryPath).PostMatchFunc(routevar.FixTreeEntryVars).BuildVarsFunc(routevar.PrepareTreeEntryRouteVars).Methods("GET").Name(RepoTreeEntry)

	repoRev.Path("/.tree-search").Methods("GET").Name(RepoTreeSearch)

	base.Path(`/people/` + routevar.Person).Methods("GET").Name(Person)

	base.Path("/users").Methods("GET").Name(Users)
	userPath := `/users/` + routevar.User
	base.Path(userPath).Methods("GET").Name(User)
	user := base.PathPrefix(userPath).Subrouter()
	user.Path("/orgs").Methods("GET").Name(UserOrgs)
	user.Path("/emails").Methods("GET").Name(UserEmails)
	user.Path("/settings").Methods("GET").Name(UserSettings)
	user.Path("/settings").Methods("PUT").Name(UserSettingsUpdate)

	orgPath := "/orgs/{OrgSpec}"
	base.Path(orgPath).Methods("GET").Name(Org)
	org := base.PathPrefix(orgPath).Subrouter()
	org.Path("/settings").Methods("GET").Name(OrgSettings)
	org.Path("/settings").Methods("PUT").Name(OrgSettingsUpdate)
	org.Path("/members").Methods("GET").Name(OrgMembers)

	base.Path("/search").Methods("GET").Name(Search)
	base.Path("/search/complete").Methods("GET").Name(SearchComplete)
	base.Path("/search/suggestions").Methods("GET").Name(SearchSuggestions)

	base.Path("/.defs").Methods("GET").Name(Defs)

	// See router_util/def_route.go for an explanation of how we match def
	// routes.
	defPath := `/.defs/` + routevar.Def
	repoRev.Path(defPath).Methods("GET").PostMatchFunc(routevar.FixDefUnitVars).BuildVarsFunc(routevar.PrepareDefRouteVars).Name(Def)
	def := repoRev.PathPrefix(defPath).PostMatchFunc(routevar.FixDefUnitVars).BuildVarsFunc(routevar.PrepareDefRouteVars).Subrouter()
	def.Path("/.refs").Methods("GET").Name(DefRefs)
	def.Path("/.examples").Methods("GET").Name(DefExamples)
	def.Path("/.authors").Methods("GET").Name(DefAuthors)
	def.Path("/.clients").Methods("GET").Name(DefClients)

	base.Path("/.units").Methods("GET").Name(Units)
	unitPath := `/.units/{UnitType}/{Unit:.*}`
	repoRev.Path(unitPath).Methods("GET").Name(Unit)

	base.Path("/markdown").Methods("POST").Name(Markdown)

	base.Path("/ext/github/webhook").Methods("POST").Name(ExtGitHubReceiveWebhook)

	if ExtraConfig != nil {
		ExtraConfig(base, user)
	}

	return base
}

// ExtraConfig, if non-nil, is called by NewAPIRouter with the
// *mux.Router after setting up routes in this package and before
// returning it. It can be used by external packages that use this API
// router and want to add additional routes to it.
var ExtraConfig func(base, user *mux.Router)
