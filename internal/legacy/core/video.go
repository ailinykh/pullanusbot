package core

import "os"

// Video ...
type Video struct {
	File
	ID       string
	Width    int
	Height   int
	Bitrate  int
	Duration int
	Codec    string
	Thumb    *Image
}

// Dispose to cleanup filesystem
func (vf *Video) Dispose() {
	os.Remove(vf.Path)
	os.Remove(vf.Thumb.Path)
}
