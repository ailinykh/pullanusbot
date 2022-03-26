package usecases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateUserStorageDecorator(primary core.IUserStorage, secondary core.IUserStorage) core.IUserStorage {
	return &UserStorageDecorator{primary, secondary}
}

type UserStorageDecorator struct {
	primary   core.IUserStorage
	secondary core.IUserStorage
}

// GetUserById is a core.IUserStorage interface implementation
func (decorator *UserStorageDecorator) GetUserById(userID int64) (*core.User, error) {
	user, err := decorator.primary.GetUserById(userID)
	if err != nil {
		return decorator.secondary.GetUserById(userID)
	}
	return user, err
}

// CreateUser is a core.IUserStorage interface implementation
func (decorator *UserStorageDecorator) CreateUser(user *core.User) error {
	err := decorator.primary.CreateUser(user)
	if err != nil {
		return decorator.secondary.CreateUser(user)
	}
	return err
}
