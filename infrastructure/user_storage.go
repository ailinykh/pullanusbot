package infrastructure

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func CreateUserStorage(dbFile string, l core.ILogger) core.IUserStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&User{})
	return &UserStorage{conn, l}
}

type UserStorage struct {
	conn *gorm.DB
	l    core.ILogger
}

// User
type User struct {
	UserID       int `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

// GetUserById is a core.IUserStorage interface implementation
func (storage *UserStorage) GetUserById(userID core.UserID) (*core.User, error) {
	var user User
	res := storage.conn.First(&user, userID)

	if res.Error != nil {
		storage.l.Error(res.Error)
		return nil, res.Error
	}
	return &core.User{
		ID:           user.UserID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode}, nil
}

// CreateUser is a core.IUserStorage interface implementation
func (storage *UserStorage) CreateUser(user *core.User) error {
	res := storage.conn.Create(User{
		UserID:       user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode,
	})
	return res.Error
}
