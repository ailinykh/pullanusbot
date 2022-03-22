package test_helpers

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateChatStorage() *FakeChatStorage {
	return &FakeChatStorage{make(map[int64]*core.Chat), nil}
}

type FakeChatStorage struct {
	chats map[int64]*core.Chat
	Err   error
}

// GetChatByID is a core.IChatStorage interface implementation
func (storage *FakeChatStorage) GetChatByID(chatID int64) (*core.Chat, error) {
	if user, ok := storage.chats[chatID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateChat is a core.IChatStorage interface implementation
func (s *FakeChatStorage) CreateChat(chatID int64, title string, type_ string, settings *core.Settings) error {
	return nil
}

// UpdateSettings is a core.IChatStorage interface implementation
func (s *FakeChatStorage) UpdateSettings(chatID int64, settings *core.Settings) error {
	chat, err := s.GetChatByID(chatID)
	if err != nil {
		return err
	}
	chat.Settings = settings
	return nil
}
