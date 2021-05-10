package core

type IBot interface {
	SendText(string) error
	SendVideo(*VideoFile, string) error
}
