package core

// IGameStorage is an abstract interface for game players and results handling
type IGameStorage interface {
	GetPlayers() ([]*User, error)
	GetRounds() ([]*Round, error)
	AddPlayer(*User) error
	AddRound(*Round) error
}
