package infrastructure

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateInMemoryChatStorage() core.IChatStorage {
	return &InMemoryChatStorage{make(map[int64]*core.Chat)}
}

type InMemoryChatStorage struct {
	cache map[int64]*core.Chat
}

// GetChatByID is a core.IChatStorage interface implementation
func (storage *InMemoryChatStorage) GetChatByID(chatID int64) (*core.Chat, error) {
	if chat, ok := storage.cache[chatID]; ok {
		return chat, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateChat is a core.IChatStorage interface implementation
func (storage *InMemoryChatStorage) CreateChat(chatID int64, title string, type_ string, settings *core.Settings) error {
	storage.cache[chatID] = &core.Chat{
		ID:       chatID,
		Title:    title,
		Type:     type_,
		Settings: settings,
	}
	return nil
}

// UpdateSettings is a core.IChatStorage interface implementation
func (storage *InMemoryChatStorage) UpdateSettings(chatID int64, settings *core.Settings) error {
	chat, err := storage.GetChatByID(chatID)
	if err != nil {
		return err
	}
	chat.Settings = settings
	return nil
}
