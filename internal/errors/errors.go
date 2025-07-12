package errvalues

import "errors"

var (
	ErrManyTasks  = errors.New("tasks limit exceed")
	ErrNoSuchTask = errors.New("there is no task with such id")
	ErrTaskFull   = errors.New("task completed")
)
