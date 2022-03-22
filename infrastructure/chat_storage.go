package infrastructure

import (
	"encoding/json"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateChatStorage is a default ChatStorage factory
func CreateChatStorage(dbFile string, l core.ILogger) *ChatStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&Chat{})
	return &ChatStorage{conn, l}
}

// ChatStorage implements core.IChatStorage interface
type ChatStorage struct {
	conn *gorm.DB
	l    core.ILogger
}

type Chat struct {
	ID        int64 `gorm:"primaryKey"`
	Title     string
	Type      string
	Settings  []byte
	CreatedAt time.Time `gorm:"autoUpdateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

// GetChatByID is a core.IChatStorage interface implementation
func (s *ChatStorage) GetChatByID(chatID int64) (*core.Chat, error) {
	var chat Chat
	err := s.conn.First(&chat, chatID).Error

	if err != nil {
		s.l.Error(err)
		return nil, err
	}

	var settings core.Settings
	err = json.Unmarshal(chat.Settings, &settings)

	if err != nil {
		s.l.Error(err)
		return nil, err
	}

	return &core.Chat{ID: chat.ID, Title: chat.Title, Type: chat.Type, Settings: &settings}, nil
}

// CreateChat is a core.IChatStorage interface implementation
func (s *ChatStorage) CreateChat(chatID int64, title string, type_ string, settings *core.Settings) error {
	data, err := json.Marshal(&settings)

	if err != nil {
		s.l.Error(err)
		return err
	}

	s.l.Infof("creating chat id: %d, title: %s, type: %s, data: %s", chatID, title, type_, data)
	chat := Chat{ID: chatID, Title: title, Type: type_, Settings: data}
	err = s.conn.Create(&chat).Error
	if err != nil {
		s.l.Error(err)
		return err
	}

	s.l.Info("chat created: %+v", chat)
	return nil
}

// UpdateSettings is a core.IChatStorage interface implementation
func (s *ChatStorage) UpdateSettings(chatID int64, settings *core.Settings) error {
	chat, err := s.GetChatByID(chatID)
	if err != nil {
		return err
	}
	chat.Settings = settings
	return s.conn.Save(&chat).Error
}
