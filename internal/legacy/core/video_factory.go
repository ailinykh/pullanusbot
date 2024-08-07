package core

// IVideoFactory retreives video file parameters from file on disk
type IVideoFactory interface {
	CreateVideo(path string) (*Video, error)
}
