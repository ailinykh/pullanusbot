package core

// IGameStorage is an abstract interface for game players and results handling
type IGameStorage interface {
	GetPlayers(ChatID) ([]*User, error)
	GetRounds(ChatID) ([]*Round, error)
	AddPlayer(ChatID, *User) error
	UpdatePlayer(ChatID, *User) error
	AddRound(ChatID, *Round) error
}
