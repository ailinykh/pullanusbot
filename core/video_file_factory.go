package core

type IVideoFileFactory interface {
	CreateVideoFile(path string) (*VideoFile, error)
}
