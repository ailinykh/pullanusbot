package core

func CreateImage(id string, fileURL string) Image {
	return Image{ID: id, FileURL: fileURL}
}

type Image struct {
	File
	ID      string
	FileURL string
}
