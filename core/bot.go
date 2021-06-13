package core

type IBot interface {
	IsPrivate() bool
	SendText(string) error
	SendPhoto(*Media) error
	SendPhotoAlbum([]*Media) error
	SendVideo(*Media) error
	SendVideoFile(*VideoFile, string) error
}
