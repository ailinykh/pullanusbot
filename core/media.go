package core

// MediaType ...
type MediaType int

const (
	// Video media type
	TVideo MediaType = iota
	// Photo media type
	TPhoto
	// Text media type
	TText
)

// Media ...
type Media struct {
	ResourceURL URL
	URL         URL
	Title       string
	Description string
	Caption     string
	Duration    int    // video only
	Codec       string // video only
	Type        MediaType
}

type ISendMediaStrategy interface {
	SendMedia([]*Media, IBot) error
}
