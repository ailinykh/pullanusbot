package core

// IVideoConverter convert Video with specified bitrate
type IVideoConverter interface {
	Convert(*Video, int) (*Video, error)
}
