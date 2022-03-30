package core

// IBot represents abstract bot interface
type IBot interface {
	Delete(*Message) error
	Edit(*Message, interface{}, ...interface{}) (*Message, error)
	SendText(string, ...interface{}) (*Message, error)
	SendImage(*Image, string) (*Message, error)
	SendAlbum([]*Image) ([]*Message, error)
	SendMedia(*Media) (*Message, error)
	SendPhotoAlbum([]*Media) ([]*Message, error)
	SendVideo(*Video, string) (*Message, error)
	IsUserMemberOfChat(*User, int64) bool
	GetCommands(int64) ([]Command, error)
	SetCommands(int64, []Command) error
}
