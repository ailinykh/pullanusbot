package core

type MediaType int

const (
	Video MediaType = iota
	Photo
	Text
)

type Media struct {
	URL     string
	Caption string
	Type    MediaType
}
