package infrastructure

// IPlayerFactory creates infrastructure Player representation
type IPlayerFactory interface {
	CreatePlayer(string) Player
}
