package core

type ISettingsProvider interface {
	GetData(ChatID, SettingKey) ([]byte, error)
	SetData(ChatID, SettingKey, []byte) error
}
