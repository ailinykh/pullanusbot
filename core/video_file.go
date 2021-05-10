package core

import "os"

type VideoFile struct {
	Width     int
	Height    int
	Bitrate   int
	Duration  int
	Codec     string
	FileName  string
	FilePath  string
	ThumbPath string
}

func (vf *VideoFile) Dispose() {
	os.Remove(vf.FilePath)
	os.Remove(vf.ThumbPath)
}
