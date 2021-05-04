package infrastructure

type IPlayerFactory interface {
	Make(string) Player
}
