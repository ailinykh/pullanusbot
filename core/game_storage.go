package core

// IGameStorage is an abstract interface for game players and results handling
type IGameStorage interface {
	GetPlayers(int64) ([]*User, error)
	GetRounds(int64) ([]*Round, error)
	AddPlayer(int64, *User) error
	AddRound(int64, *Round) error
}
