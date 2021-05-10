package core

type IGameStorage interface {
	GetPlayers() ([]*User, error)
	GetRounds() ([]*Round, error)
	AddPlayer(*User) error
	AddRound(*Round) error
}
