package sourcegraph

import (
	"errors"
	"fmt"
	"time"

	"strconv"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
	"sourcegraph.com/sourcegraph/srclib/db_common"
)

// BuildsService communicates with the build-related endpoints in the
// Sourcegraph API.
type BuildsService interface {
	// Get fetches a build.
	Get(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error)

	// List builds.
	List(opt *BuildListOptions) ([]*Build, Response, error)

	// ListByRepo lists builds for a repository.
	ListByRepo(repo RepoSpec, opt *BuildListByRepoOptions) ([]*Build, Response, error)

	// Create a new build. The build will run asynchronously (Create does not
	// wait for it to return. To monitor the build's status, use Get.)
	Create(repo RepoSpec, opt *BuildCreateOptions) (*Build, Response, error)

	// Update updates information about a build and returns the build
	// after the update has been applied.
	Update(build BuildSpec, info BuildUpdate) (*Build, Response, error)

	// ListBuildTasks lists the tasks associated with a build.
	ListBuildTasks(build BuildSpec, opt *BuildTaskListOptions) ([]*BuildTask, Response, error)

	// CreateTasks creates tasks associated with a build and returns
	// them with their TID fields set.
	CreateTasks(build BuildSpec, tasks []*BuildTask) ([]*BuildTask, Response, error)

	// UpdateTask updates a task associated with a build.
	UpdateTask(task TaskSpec, info TaskUpdate) (*BuildTask, Response, error)

	// GetLog gets log entries associated with a build.
	GetLog(build BuildSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error)

	// GetTaskLog gets log entries associated with a task.
	GetTaskLog(task TaskSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error)
}

type buildsService struct {
	client *Client
}

var _ BuildsService = &buildsService{}

type BuildSpec struct {
	BID int64
}

func (s *BuildSpec) RouteVars() map[string]string {
	return map[string]string{"BID": fmt.Sprintf("%d", s.BID)}
}

type TaskSpec struct {
	BuildSpec
	TaskID int64
}

func (s *TaskSpec) RouteVars() map[string]string {
	v := s.BuildSpec.RouteVars()
	v["TaskID"] = fmt.Sprintf("%d", s.TaskID)
	return v
}

// A Build represents a scheduled, completed, or failed repository analysis and
// import job.
type Build struct {
	BID         int64 `json:",omitempty"`
	Repo        int
	CreatedAt   time.Time          `db:"created_at"`
	StartedAt   db_common.NullTime `db:"started_at"`
	EndedAt     db_common.NullTime `db:"ended_at"`
	HeartbeatAt db_common.NullTime `db:"heartbeat_at"`
	Success     bool               `json:",omitempty"`
	Failure     bool               `json:",omitempty"`

	// Killed is true if this build's worker didn't exit on its own
	// accord. It is generally set when no heartbeat has been received
	// within a certain interval. If Killed is true, then Failure must
	// also always be set to true.
	Killed bool `json:",omitempty"`

	// Host is the hostname of the machine that is working on this build.
	Host string `json:",omitempty"`

	Purged bool // whether the build's data (defs/refs/etc.) has been purged

	BuildConfig

	// RepoURI is populated (as a convenience) in results by Get and List but
	// should not be set when creating builds (it will be ignored).
	RepoURI *string `db:"repo_uri" json:",omitempty"`
}

func (b *Build) Spec() BuildSpec { return BuildSpec{BID: b.BID} }

// IDString returns a succinct string that uniquely identifies this build.
func (b BuildSpec) IDString() string { return buildIDString(b.BID) }

func buildIDString(bid int64) string { return "B" + strconv.FormatInt(bid, 36) }

// A BuildTask represents an individual step of a build.
type BuildTask struct {
	// TaskID is the unique ID of this task. It is unique over all
	// tasks, not just tasks in the same build.
	TaskID int64 `json:",omitempty"`

	// BID is the build that this task is a part of.
	BID int64

	// UnitType is the srclib source unit type of the source unit that
	// this task is associated with.
	UnitType string `db:"unit_type" json:",omitempty"`

	// Unit is the srclib source unit name of the source unit that
	// this task is associated with.
	Unit string `json:",omitempty"`

	// Op is the srclib toolchain operation (graph, depresolve, etc.) that this
	// task performs.
	Op string `json:",omitempty"`

	// Order is the order in which this task is performed, relative to other
	// tasks in the same build. Lower-number-ordered tasks are built first.
	// Multiple tasks may have the same order.
	Order int `json:",omitempty"`

	// CreatedAt is when this task was initially created.
	CreatedAt time.Time `db:"created_at"`

	// StartedAt is when this task's execution began.
	StartedAt db_common.NullTime `db:"started_at" json:",omitempty"`

	// EndedAt is when this task's execution ended (whether because it
	// succeeded or failed).
	EndedAt db_common.NullTime `db:"ended_at" json:",omitempty"`

	// Queue is whether this task should be performed by queue task
	// remote workers on the central server. If true, then it will be
	// performed remotely. If false, it should be performed locally by
	// the process that created this task.
	//
	// For example, import tasks are queued because they are performed
	// by the remote server, not the local "src" process running on
	// the builders.
	Queue bool

	// Success is whether this task's execution succeeded.
	Success bool `json:",omitempty"`

	// Failure is whether this task's execution failed.
	Failure bool `json:",omitempty"`
}

func (t *BuildTask) Spec() TaskSpec {
	return TaskSpec{BuildSpec: BuildSpec{BID: t.BID}, TaskID: t.TaskID}
}

// IDString returns a succinct string that uniquely identifies this build task.
func (t TaskSpec) IDString() string {
	return buildIDString(t.BID) + "-T" + strconv.FormatInt(t.TaskID, 36)
}

// BuildConfig configures a repository build.
type BuildConfig struct {
	// Import is whether to import the build data into the database when the
	// build is complete. The data must be imported for Sourcegraph's web app or
	// API to use it, except that unimported build data is available through the
	// BuildData service. (TODO(sqs): BuildData isn't yet implemented.)
	Import bool

	// Queue is whether this build should be enqueued. If enqueued, any worker
	// may begin running this build. If not enqueued, it is up to the client to
	// run the build and update it accordingly.
	Queue bool

	// UseCache is whether to use cached build data files. If false, the
	// .sourcegraph-data directory will be wiped out before the build begins.
	//
	// Regardless of the value of UseCache, the build data files will be
	// uploaded to the central cache after the build ends.
	UseCache bool `db:"use_cache"`

	// Priority of the build in the queue (higher numbers mean the build is
	// dequeued sooner).
	Priority int

	// CommitID is the full resolved commit ID to build.
	CommitID string `db:"commit_id"`
}

type BuildCreateOptions struct {
	BuildConfig

	// Force creation of build (if false, the build will not be
	// created if a build for the same repository and commit ID
	// exists).
	//
	// TODO(bliu): test this
	Force bool
}

var ErrBuildNotFound = errors.New("build not found")

type BuildGetOptions struct{}

func (s *buildsService) Get(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error) {
	url, err := s.client.url(router.Build, build.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var build_ *Build
	resp, err := s.client.Do(req, &build_)
	if err != nil {
		return nil, resp, err
	}

	return build_, nil, nil
}

type BuildListOptions struct {
	Queued    bool `url:",omitempty"`
	Active    bool `url:",omitempty"`
	Ended     bool `url:",omitempty"`
	Succeeded bool `url:",omitempty"`
	Failed    bool `url:",omitempty"`

	Purged bool `url:",omitempty"`

	Sort      string `url:",omitempty"`
	Direction string `url:",omitempty"`

	ListOptions
}

func (s *buildsService) List(opt *BuildListOptions) ([]*Build, Response, error) {
	url, err := s.client.url(router.Builds, nil, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var builds []*Build
	resp, err := s.client.Do(req, &builds)
	if err != nil {
		return nil, resp, err
	}

	return builds, resp, nil
}

type BuildListByRepoOptions struct {
	BuildListOptions
	Rev string `url:",omitempty"`
}

func (s *buildsService) ListByRepo(repo RepoSpec, opt *BuildListByRepoOptions) ([]*Build, Response, error) {
	url, err := s.client.url(router.RepoBuilds, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var builds []*Build
	resp, err := s.client.Do(req, &builds)
	if err != nil {
		return nil, resp, err
	}

	return builds, resp, nil
}

func (s *buildsService) Create(repo RepoSpec, opt *BuildCreateOptions) (*Build, Response, error) {
	url, err := s.client.url(router.RepoBuildsCreate, repo.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", url.String(), opt)
	if err != nil {
		return nil, nil, err
	}

	var build *Build
	resp, err := s.client.Do(req, &build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, nil
}

type BuildTaskListOptions struct{ ListOptions }

func (s *buildsService) ListBuildTasks(build BuildSpec, opt *BuildTaskListOptions) ([]*BuildTask, Response, error) {
	url, err := s.client.url(router.BuildTasks, build.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*BuildTask
	resp, err := s.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, nil
}

// A BuildUpdate contains updated information to update on an existing
// build.
type BuildUpdate struct {
	StartedAt   *time.Time
	EndedAt     *time.Time
	HeartbeatAt *time.Time
	Host        *string
	Success     *bool
	Purged      *bool
	Failure     *bool
	Killed      *bool
	Priority    *int
}

func (s *buildsService) Update(build BuildSpec, info BuildUpdate) (*Build, Response, error) {
	url, err := s.client.url(router.BuildUpdate, build.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), info)
	if err != nil {
		return nil, nil, err
	}

	var updated *Build
	resp, err := s.client.Do(req, &updated)
	if err != nil {
		return nil, resp, err
	}

	return updated, resp, nil
}

func (s *buildsService) CreateTasks(build BuildSpec, tasks []*BuildTask) ([]*BuildTask, Response, error) {
	url, err := s.client.url(router.BuildTasksCreate, build.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", url.String(), tasks)
	if err != nil {
		return nil, nil, err
	}

	var created []*BuildTask
	resp, err := s.client.Do(req, &created)
	if err != nil {
		return nil, resp, err
	}

	return created, resp, nil
}

// A TaskUpdate contains updated information to update on an existing
// task.
type TaskUpdate struct {
	StartedAt *time.Time
	EndedAt   *time.Time
	Success   *bool
	Failure   *bool
}

func (s *buildsService) UpdateTask(task TaskSpec, info TaskUpdate) (*BuildTask, Response, error) {
	url, err := s.client.url(router.BuildTaskUpdate, task.RouteVars(), nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("PUT", url.String(), info)
	if err != nil {
		return nil, nil, err
	}

	var updated *BuildTask
	resp, err := s.client.Do(req, &updated)
	if err != nil {
		return nil, resp, err
	}

	return updated, resp, nil
}

// BuildGetLogOptions specifies options for build log API methods.
type BuildGetLogOptions struct {
	// MinID indicates that only log entries whose monotonically increasing ID
	// is greater than MinID should be returned.
	//
	// To "tail -f" or watch a log for updates, set each subsequent request's
	// MinID to the MaxID of the previous request.
	MinID string
}

type LogEntries struct {
	MaxID   string
	Entries []string
}

func (s *buildsService) GetLog(build BuildSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error) {
	url, err := s.client.url(router.BuildLog, build.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var entries *LogEntries
	resp, err := s.client.Do(req, &entries)
	if err != nil {
		return nil, resp, err
	}

	return entries, resp, nil
}

func (s *buildsService) GetTaskLog(task TaskSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error) {
	url, err := s.client.url(router.BuildTaskLog, task.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var entries *LogEntries
	resp, err := s.client.Do(req, &entries)
	if err != nil {
		return nil, resp, err
	}

	return entries, resp, nil
}

type MockBuildsService struct {
	Get_            func(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error)
	List_           func(opt *BuildListOptions) ([]*Build, Response, error)
	ListByRepo_     func(repo RepoSpec, opt *BuildListByRepoOptions) ([]*Build, Response, error)
	Create_         func(repo RepoSpec, opt *BuildCreateOptions) (*Build, Response, error)
	ListBuildTasks_ func(build BuildSpec, opt *BuildTaskListOptions) ([]*BuildTask, Response, error)
	Update_         func(build BuildSpec, info BuildUpdate) (*Build, Response, error)
	CreateTasks_    func(build BuildSpec, tasks []*BuildTask) ([]*BuildTask, Response, error)
	UpdateTask_     func(task TaskSpec, info TaskUpdate) (*BuildTask, Response, error)
	GetLog_         func(build BuildSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error)
	GetTaskLog_     func(task TaskSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error)
}

var _ BuildsService = MockBuildsService{}

func (s MockBuildsService) Get(build BuildSpec, opt *BuildGetOptions) (*Build, Response, error) {
	if s.Get_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.Get_(build, opt)
}

func (s MockBuildsService) List(opt *BuildListOptions) ([]*Build, Response, error) {
	if s.List_ == nil {
		return nil, nil, nil
	}
	return s.List_(opt)
}

func (s MockBuildsService) ListByRepo(repo RepoSpec, opt *BuildListByRepoOptions) ([]*Build, Response, error) {
	if s.ListByRepo_ == nil {
		return nil, &HTTPResponse{}, nil
	}
	return s.ListByRepo_(repo, opt)
}

func (s MockBuildsService) Create(repo RepoSpec, opt *BuildCreateOptions) (*Build, Response, error) {
	if s.Create_ == nil {
		return nil, nil, nil
	}
	return s.Create_(repo, opt)
}

func (s MockBuildsService) ListBuildTasks(build BuildSpec, opt *BuildTaskListOptions) ([]*BuildTask, Response, error) {
	if s.ListBuildTasks_ == nil {
		return nil, nil, nil
	}
	return s.ListBuildTasks_(build, opt)
}

func (s MockBuildsService) Update(build BuildSpec, info BuildUpdate) (*Build, Response, error) {
	return s.Update_(build, info)
}

func (s MockBuildsService) CreateTasks(build BuildSpec, tasks []*BuildTask) ([]*BuildTask, Response, error) {
	return s.CreateTasks_(build, tasks)
}

func (s MockBuildsService) UpdateTask(task TaskSpec, info TaskUpdate) (*BuildTask, Response, error) {
	return s.UpdateTask_(task, info)
}

func (s MockBuildsService) GetLog(build BuildSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error) {
	if s.GetLog_ == nil {
		return nil, nil, nil
	}
	return s.GetLog_(build, opt)
}

func (s MockBuildsService) GetTaskLog(task TaskSpec, opt *BuildGetLogOptions) (*LogEntries, Response, error) {
	if s.GetTaskLog_ == nil {
		return nil, nil, nil
	}
	return s.GetTaskLog_(task, opt)
}
