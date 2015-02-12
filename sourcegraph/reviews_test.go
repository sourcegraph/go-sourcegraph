package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

func TestReviewsService_ListTasks(t *testing.T) {
	setup()
	defer teardown()

	want := []*ReviewTask{{}, {}}
	reviewSpec := ReviewSpec{Repo: RepoSpec{URI: "r.com/x"}, Number: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.ReviewTasks, reviewSpec.RouteVars()), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"PerPage": "1",
			"Page":    "2",
		})

		writeJSON(w, want)
	})

	tasks, _, err := client.Reviews.ListTasks(
		reviewSpec,
		&ReviewListTasksOptions{
			ListOptions: ListOptions{PerPage: 1, Page: 2},
		},
	)
	if err != nil {
		t.Errorf("Reviews.ListTasks returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normReviewTask(want...)
	if !reflect.DeepEqual(tasks, want) {
		t.Errorf("Reviews.ListTasks returned %+v, want %+v", tasks, want)
	}
}

func TestReviewsService_ListTasksByRepo(t *testing.T) {
	setup()
	defer teardown()

	want := []*ReviewTask{{}, {}}
	repoSpec := RepoSpec{URI: "x.com/r"}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoReviewTasks, repoSpec.RouteVars()), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"PerPage": "1",
			"Page":    "2",
		})

		writeJSON(w, want)
	})

	tasks, _, err := client.Reviews.ListTasksByRepo(
		repoSpec,
		&ReviewListTasksByRepoOptions{
			ListOptions: ListOptions{PerPage: 1, Page: 2},
		},
	)
	if err != nil {
		t.Errorf("Reviews.ListTasksByRepo returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normReviewTask(want...)
	if !reflect.DeepEqual(tasks, want) {
		t.Errorf("Reviews.ListTasksByRepo returned %+v, want %+v", tasks, want)
	}
}

func TestReviewsService_ListTasksByUser(t *testing.T) {
	setup()
	defer teardown()

	want := []*ReviewTask{{}, {}}
	userSpec := UserSpec{Login: "u"}

	var called bool
	mux.HandleFunc(urlPath(t, router.UserReviewTasks, userSpec.RouteVars()), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"PerPage": "1",
			"Page":    "2",
		})

		writeJSON(w, want)
	})

	tasks, _, err := client.Reviews.ListTasksByUser(
		userSpec,
		&ReviewListTasksByUserOptions{
			ListOptions: ListOptions{PerPage: 1, Page: 2},
		},
	)
	if err != nil {
		t.Errorf("Reviews.ListTasksByUser returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	normReviewTask(want...)
	if !reflect.DeepEqual(tasks, want) {
		t.Errorf("Reviews.ListTasksByUser returned %+v, want %+v", tasks, want)
	}
}

func normReviewTask(tasks ...*ReviewTask) {
	for _, t := range tasks {
		t.CreatedAt = t.CreatedAt.In(time.UTC)
	}
}
