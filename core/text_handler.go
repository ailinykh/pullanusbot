package core

type ITextHandler interface {
	HandleText(string, IBot) error
}
