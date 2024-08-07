package core

type Command struct {
	Text        string
	Description string
}

func DefaultCommands() []Command {
	return []Command{{Text: "help", Description: ""}}
}

type ICommandService interface {
	EnableCommands(ChatID, []Command, IBot) error
	DisableCommands(ChatID, []Command, IBot) error
}
