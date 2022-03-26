package usecases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateChatStorageDecorator(cache core.IChatStorage, db core.IChatStorage) core.IChatStorage {
	return &ChatStorageDecorator{cache, db}
}

type ChatStorageDecorator struct {
	cache core.IChatStorage
	db    core.IChatStorage
}

// GetChatByID is a core.IChatStorage interface implementation
func (decorator *ChatStorageDecorator) GetChatByID(chatID int64) (*core.Chat, error) {
	chat, err := decorator.cache.GetChatByID(chatID)
	if err != nil {
		chat, err := decorator.db.GetChatByID(chatID)
		if err != nil {
			return nil, err
		}
		_ = decorator.cache.CreateChat(chat.ID, chat.Title, chat.Type, chat.Settings)
		return chat, nil
	}
	return chat, nil
}

// CreateChat is a core.IChatStorage interface implementation
func (decorator *ChatStorageDecorator) CreateChat(chatID int64, title string, type_ string, settings *core.Settings) error {
	_ = decorator.cache.CreateChat(chatID, title, type_, settings)
	return decorator.db.CreateChat(chatID, title, type_, settings)
}

// UpdateSettings is a core.IChatStorage interface implementation
func (decorator *ChatStorageDecorator) UpdateSettings(chatID int64, settings *core.Settings) error {
	_ = decorator.cache.UpdateSettings(chatID, settings)
	return decorator.db.UpdateSettings(chatID, settings)
}
