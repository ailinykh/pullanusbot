package core

type ITaskFactory interface {
	NewTask(string) ITask
}

type ITask interface {
	Perform([]byte, chan []byte) error
}
