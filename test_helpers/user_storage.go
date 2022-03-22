package test_helpers

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateUserStorage() *FakeUserStorage {
	return &FakeUserStorage{make(map[int]*core.User), nil}
}

type FakeUserStorage struct {
	Users map[core.UserID]*core.User
	Err   error
}

// GetUserById is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) GetUserById(userID core.UserID) (*core.User, error) {
	if user, ok := storage.Users[userID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateUser is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) CreateUser(user *core.User) error {
	storage.Users[user.ID] = user
	return nil
}
