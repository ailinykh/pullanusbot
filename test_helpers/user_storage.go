package test_helpers

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateUserStorage() *FakeUserStorage {
	return &FakeUserStorage{make(map[int]*core.User), nil}
}

type FakeUserStorage struct {
	users map[core.UserID]*core.User
	Err   error
}

// GetUserById is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) GetUserById(userID core.UserID) (*core.User, error) {
	if user, ok := storage.users[userID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user with id %d not found", userID)
}

// CreateUser is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) CreateUser(user *core.User) error {
	storage.users[user.ID] = user
	return nil
}
