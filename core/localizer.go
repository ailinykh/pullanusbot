package core

// ILocalizer for localization
type ILocalizer interface {
	I18n(string, ...interface{}) string
	AllKeys() []string
}
