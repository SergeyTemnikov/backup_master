package model

type TaskState string

const (
	TaskIdle    TaskState = "idle"
	TaskRunning TaskState = "running"
	TaskSuccess TaskState = "success"
	TaskError   TaskState = "error"
)
