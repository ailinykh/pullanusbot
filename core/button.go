package core

type Keyboard = [][]*Button

type IButtonHandler interface {
	GetButtonIds() []string
	ButtonPressed(string, *Message, IBot) error
}

type Button struct {
	ID   string
	Text string
}
