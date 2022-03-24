package core

type IUserStorage interface {
	GetUserById(int64) (*User, error)
	CreateUser(*User) error
}
