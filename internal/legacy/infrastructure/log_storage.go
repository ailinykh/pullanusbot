package infrastructure

import (
	"time"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewLogStorage(dbFile string, l core.Logger) *LogStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&LogRecord{})
	return &LogStorage{conn, l}
}

type LogStorage struct {
	conn *gorm.DB
	l    core.Logger
}

type LogRecord struct {
	ID        int64 `gorm:"primaryKey"`
	Method    string
	Data      string
	CreatedAt time.Time `gorm:"autoUpdateTime"`
}

func (s *LogStorage) LogRecord(method string, data string) error {
	logRecord := LogRecord{Method: method, Data: data}
	err := s.conn.Create(&logRecord).Error
	if err != nil {
		s.l.Error(err)
	}
	return nil
}
