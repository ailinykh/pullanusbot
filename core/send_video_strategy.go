package core

type ISendVideoStrategy interface {
	SendVideo(*Video, string, IBot) error
}
