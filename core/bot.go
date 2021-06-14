package core

type IBot interface {
	Delete(*Message) error
	SendText(string) error
	SendImage(*Image) (*Message, error)
	SendAlbum([]*Image) ([]*Message, error)
	SendPhoto(*Media) error
	SendPhotoAlbum([]*Media) error
	SendVideo(*Media) error
	SendVideoFile(*VideoFile, string) error
}