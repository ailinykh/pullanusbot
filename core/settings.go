package core

type Settings struct {
	FaggotGameCommandsEnabled bool `json:"faggot_game_enabled"`
	RemoveSourceOnSucccess    bool `json:"remove_source_on_success"`
}

func DefaultSettings() Settings {
	return Settings{false, true}
}

type ISettingsStorage interface {
	GetSettings(int64) (*Settings, error)
	SetSettings(int64, *Settings) error
}
