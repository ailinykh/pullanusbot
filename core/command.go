package core

type Command struct {
	Text        string
	Description string
}

func DefaultCommands() []Command {
	return []Command{{Text: "help", Description: ""}}
}

type ICommandService interface {
	EnableCommands(int64, []Command) error
	DisableCommands(int64, []Command) error
}
