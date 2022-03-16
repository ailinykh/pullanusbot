package test_helpers

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateSettingsStorage() *FakeSettingsStorage {
	return &FakeSettingsStorage{make(map[int64]*core.Settings), nil}
}

type FakeSettingsStorage struct {
	Data map[int64]*core.Settings
	Err  error
}

// GetSettings is a core.ISettingsStorage interface implementation
func (s *FakeSettingsStorage) GetSettings(chatID int64) (*core.Settings, error) {
	if settings, ok := s.Data[chatID]; ok {
		return settings, nil
	}

	settings := core.DefaultSettings()
	s.Data[chatID] = &settings
	return &settings, nil
}

// SetSettings is a core.ISettingsStorage interface implementation
func (s *FakeSettingsStorage) SetSettings(chatID int64, settings *core.Settings) error {
	s.Data[chatID] = settings
	return nil
}
