package infrastructure

import (
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
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

	return &core.Chat{ID: chat.ID, Title: chat.Title, Type: chat.Type}, nil
}

// CreateChat is a core.IChatStorage interface implementation
func (s *ChatStorage) CreateChat(chatID int64, title string, type_ string) error {
	chat := Chat{ID: chatID, Title: title, Type: type_}
	err := s.conn.Create(&chat).Error
	if err != nil {
		s.l.Error(err)
		return err
	}

	s.l.Infof("chat created: {%d %s %s}", chat.ID, chat.Title, chat.Type)
	return nil
}
