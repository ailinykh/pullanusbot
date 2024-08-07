package core

type ISettingsProvider interface {
	GetData(ChatID, SettingKey) ([]byte, error)
	SetData(ChatID, SettingKey, []byte) error
}

type IBoolSettingProvider interface {
	GetBool(ChatID, SettingKey) bool
	SetBool(ChatID, SettingKey, bool) error
}
