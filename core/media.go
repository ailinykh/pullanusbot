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
	URL      string
	Caption  string
	Duration int
	Codec    string // only video
	Type     MediaType
}
