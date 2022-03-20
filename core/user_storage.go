package core

type IUserStorage interface {
	GetUserById(UserID) (*User, error)
	CreateUser(*User) error
}
