package infrastructure

import (
	"encoding/json"
	"errors"

	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateSettingsStorage is a default SettingsStorage factory
func CreateSettingsStorage(dbFile string, l core.ILogger) *SettingsStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&Settings{})
	return &SettingsStorage{conn, l}
}

// SettingsStorage implements core.ISettingsStorage interface
type SettingsStorage struct {
	conn *gorm.DB
	l    core.ILogger
}

// GetSettings is a core.ISettingsStorage interface implementation
func (s *SettingsStorage) GetSettings(chatID int64) (*core.Settings, error) {
	defalt := core.DefaultSettings()
	data, err := json.Marshal(&defalt)

	if err != nil {
		s.l.Error(err)
		return nil, err
	}

	settings := &Settings{ChatID: chatID, Data: data}
	res := s.conn.Where("chat_id = ?", chatID).FirstOrCreate(settings)

	if res.Error != nil {
		s.l.Error(res.Error)
		return nil, res.Error
	}

	s.l.Infof("get settings %d %s", chatID, string(settings.Data))
	return makeSettings(settings.Data)
}

// SetSettings is a core.ISettingsStorage interface implementation
func (s *SettingsStorage) SetSettings(chatID int64, settings *core.Settings) error {
	data, err := json.Marshal(settings)

	if err != nil {
		s.l.Error(err)
		return err
	}

	sett := &Settings{ChatID: chatID}
	res := s.conn.First(&sett, chatID)
	sett.Data = data

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		s.l.Infof("creating settings %d %s", chatID, string(data))
		res = s.conn.Create(&sett)
	} else {
		s.l.Infof("updating settings %d %s", chatID, string(data))
		res = s.conn.Save(&sett)
	}

	return res.Error
}

func makeSettings(data []byte) (*core.Settings, error) {
	var settings *core.Settings
	err := json.Unmarshal(data, &settings)

	if err != nil {
		return nil, err
	}

	return settings, nil
}
