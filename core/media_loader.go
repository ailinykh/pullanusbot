package core

// URL ...
type URL = string

// IMediaFactory creates Media from URL
type IMediaFactory interface {
	CreateMedia(URL) ([]*Media, error)
}
