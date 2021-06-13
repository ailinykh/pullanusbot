package core

type URL = string

type IMediaFactory interface {
	CreateMedia(URL, *User) ([]*Media, error)
}
