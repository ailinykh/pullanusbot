package core

type Settings struct {
	FaggotGameCommandsEnabled bool     `json:"faggot_game_enabled"`
	Payload                   []string `json:"payload"`
	RemoveSourceOnSucccess    bool     `json:"remove_source_on_success"`
}

func DefaultSettings() Settings {
	return Settings{false, []string{}, true}
}
