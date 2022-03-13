package core

type Keyboard = [][]*Button

type IButtonHandler interface {
	AllButtons() []*Button
	ButtonPressed(string, *Message, IBot) error
}

type Button struct {
	ID   string
	Text string
}
