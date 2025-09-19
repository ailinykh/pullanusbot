package core

import "fmt"

type ITaskFactory interface {
	NewTask(string) ITask
}

type ITask interface {
	Perform([]byte, chan []byte) error
}

type TaskMock struct{}

func (TaskMock) Perform([]byte, chan []byte) error {
	return fmt.Errorf("not implemented")
}
