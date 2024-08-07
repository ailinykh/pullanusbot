package core

// ILocalizer for localization
type ILocalizer interface {
	I18n(string, string, ...interface{}) string
	AllKeys() []string
}
