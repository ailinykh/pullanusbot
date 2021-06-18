package core

import "os"

// VideoFile ...
type VideoFile struct {
	File
	Width     int
	Height    int
	Bitrate   int
	Duration  int
	Codec     string
	ThumbPath string
}

// Dispose to cleanup filesystem
func (vf *VideoFile) Dispose() {
	os.Remove(vf.Path)
	os.Remove(vf.ThumbPath)
}
