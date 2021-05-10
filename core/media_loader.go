package core

type IMediaLoader interface {
	Load(string, *User) ([]*Media, error)
}
