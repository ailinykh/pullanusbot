package core

// CreateImage is an Image factory
func CreateImage(id string, fileURL string) Image {
	return Image{ID: id, FileURL: fileURL}
}

// Image represents remote image file that can be also downloaded
type Image struct {
	File
	ID      string
	FileURL string
}
