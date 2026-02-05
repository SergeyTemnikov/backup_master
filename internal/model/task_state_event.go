package model

type TaskStateEvent struct {
	TaskID int64
	State  TaskState
}
