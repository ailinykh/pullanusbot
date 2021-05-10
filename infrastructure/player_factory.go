package infrastructure

type IPlayerFactory interface {
	CreatePlayer(string) Player
}
