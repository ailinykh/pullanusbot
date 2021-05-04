package core

type ILocalizer interface {
	I18n(string, ...interface{}) string
	AllKeys() []string
}
