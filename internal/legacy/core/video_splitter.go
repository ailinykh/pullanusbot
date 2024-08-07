package core

// IVideoSplitter convert Video with specified bitrate
type IVideoSplitter interface {
	Split(*Video, int) ([]*Video, error)
}
