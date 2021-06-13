package core

type IBot interface {
	SendText(string) error
	SendPhoto(*Media) error
	SendPhotoAlbum([]*Media) error
	SendVideo(*Media) error
	SendVideoFile(*VideoFile, string) error
}
