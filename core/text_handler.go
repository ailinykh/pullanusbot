package core

type ITextHandler interface {
	HandleText(string, *User, IBot) error
}
