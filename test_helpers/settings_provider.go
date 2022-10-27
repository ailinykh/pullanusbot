package test_helpers

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSettingsProvider() core.ISettingsProvider {
	return &FakeSettingsProvider{make(map[int64]map[string][]byte), nil}
}

type FakeSettingsProvider struct {
	Data map[core.ChatID]map[string][]byte
	Err  error
}

// GetSettings is a core.ISettingsProvider interface implementation
func (s *FakeSettingsProvider) GetData(chatID core.ChatID, key string) ([]byte, error) {
	if chat, ok := s.Data[chatID]; ok {
		if settings, ok := chat[key]; ok {
			return settings, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

// SetSettings is a core.ISettingsProvider interface implementation
func (s *FakeSettingsProvider) SetData(chatID core.ChatID, key string, data []byte) error {
	if _, ok := s.Data[chatID]; !ok {
		s.Data[chatID] = map[string][]byte{}
	}
	s.Data[chatID][key] = data
	return nil
}
