package core

// IVideoConverter convert Video with specified bitrate
type IVideoConverter interface {
	GetCodec(string) string
	Convert(*Video, int) (*Video, error)
}
