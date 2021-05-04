package core

type IGameStorage interface {
	GetPlayers() ([]Player, error)
	GetRounds() ([]Round, error)
	AddPlayer(Player) error
	AddRound(Round) error
}
