package sourcegraph

type Checklist struct {
	TodoItems, DoneItems []string

	Todo int // number of tasks to be done (unchecked)
	Done int // number of tasks that are done (checked)
}
