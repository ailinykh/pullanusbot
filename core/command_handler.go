package core

type ICommandHandler interface {
	HandleCommand(string, IBot) error
}
