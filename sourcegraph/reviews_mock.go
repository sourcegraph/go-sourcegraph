package sourcegraph

type MockReviewsService struct {
	ListTasks_       func(rv ReviewSpec, opt *ReviewListTasksOptions) ([]*ReviewTask, Response, error)
	ListTasksByRepo_ func(repo RepoSpec, opt *ReviewListTasksByRepoOptions) ([]*ReviewTask, Response, error)
	ListTasksByUser_ func(user UserSpec, opt *ReviewListTasksByUserOptions) ([]*ReviewTask, Response, error)
}

func (s MockReviewsService) ListTasks(rv ReviewSpec, opt *ReviewListTasksOptions) ([]*ReviewTask, Response, error) {
	return s.ListTasks_(rv, opt)
}

func (s MockReviewsService) ListTasksByRepo(repo RepoSpec, opt *ReviewListTasksByRepoOptions) ([]*ReviewTask, Response, error) {
	return s.ListTasksByRepo_(repo, opt)
}

func (s MockReviewsService) ListTasksByUser(user UserSpec, opt *ReviewListTasksByUserOptions) ([]*ReviewTask, Response, error) {
	return s.ListTasksByUser_(user, opt)
}
