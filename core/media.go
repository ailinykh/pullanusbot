package core

// MediaType ...
type MediaType int

const (
	// Video media type
	Video MediaType = iota
	// Photo media type
	Photo
	// Text media type
	Text
)

// Media ...
type Media struct {
	URL      string
	Caption  string
	Duration int
	Type     MediaType
}
