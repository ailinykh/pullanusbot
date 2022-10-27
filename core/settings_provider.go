package core

type ISettingsProvider interface {
	GetData(ChatID, string) ([]byte, error)
	SetData(ChatID, string, []byte) error
}
