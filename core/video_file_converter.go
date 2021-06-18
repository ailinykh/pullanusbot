package core

// IVideoFileConverter convert VideoFile with specified bitrate
type IVideoFileConverter interface {
	Convert(*VideoFile, int) (*VideoFile, error)
}
