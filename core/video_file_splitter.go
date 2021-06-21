package core

// IVideoFileSplitter convert VideoFile with specified bitrate
type IVideoFileSplitter interface {
	Split(*VideoFile, int) ([]*VideoFile, error)
}
