package core

// URL ...
type URL = string

// MediaType ...
type MediaType int

const (
	// Video media type
	TVideo MediaType = iota
	// Photo media type
	TPhoto
	// Text media type
	TText
	// Audio media type
	TAudio
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
	Size        int
	Type        MediaType
}

// IMediaFactory creates Media from URL
type IMediaFactory interface {
	CreateMedia(URL) ([]*Media, error)
}

type ISendMediaStrategy interface {
	SendMedia([]*Media, IBot) error
}
