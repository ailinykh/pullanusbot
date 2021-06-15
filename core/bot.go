package core

type IBot interface {
	Delete(*Message) error
	SendText(string, ...interface{}) (*Message, error)
	SendImage(*Image) (*Message, error)
	SendAlbum([]*Image) ([]*Message, error)
	SendMedia(*Media) (*Message, error)
	SendPhotoAlbum([]*Media) ([]*Message, error)
	SendVideoFile(*VideoFile, string) (*Message, error)
}
