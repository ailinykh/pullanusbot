package core

type Logger interface {
	Debug(...interface{})
	Error(...interface{})
	Info(...interface{})
	Warn(...interface{})
}
