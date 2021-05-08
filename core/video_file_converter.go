package core

type IVideoFileConverter interface {
	Convert(*VideoFile, int) (*VideoFile, error)
}
