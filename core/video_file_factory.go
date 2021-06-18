package core

// IVideoFileFactory retreives video file parameters from file on disk
type IVideoFileFactory interface {
	CreateVideoFile(path string) (*VideoFile, error)
}
