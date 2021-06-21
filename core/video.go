package core

import "os"

// Video ...
type Video struct {
	File
	Width     int
	Height    int
	Bitrate   int
	Duration  int
	Codec     string
	ThumbPath string
}

// Dispose to cleanup filesystem
func (vf *Video) Dispose() {
	os.Remove(vf.Path)
	os.Remove(vf.ThumbPath)
}
