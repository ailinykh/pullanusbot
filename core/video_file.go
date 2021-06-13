package core

import "os"

type VideoFile struct {
	File
	Width     int
	Height    int
	Bitrate   int
	Duration  int
	Codec     string
	ThumbPath string
}

func (vf *VideoFile) Dispose() {
	os.Remove(vf.Path)
	os.Remove(vf.ThumbPath)
}
