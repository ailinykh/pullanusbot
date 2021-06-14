package core

type IBot interface {
	Delete(*Message) error
	SendText(string) (*Message, error)
	SendImage(*Image) (*Message, error)
	SendAlbum([]*Image) ([]*Message, error)
	SendPhoto(*Media) (*Message, error)
	SendPhotoAlbum([]*Media) ([]*Message, error)
	SendVideo(*Media) (*Message, error)
	SendVideoFile(*VideoFile, string) error
}
