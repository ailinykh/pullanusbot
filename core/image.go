package core

// Image represents remote image file that can be also downloaded
type Image struct {
	File
	ID      string
	FileURL string
	Width   int
	Height  int
}
