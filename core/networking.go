package core

// IFileDownloader turns URL to File
type IFileDownloader interface {
	Download(URL) (*File, error)
}

// IFileUploader turns File to URL
type IFileUploader interface {
	Upload(*File) (URL, error)
}

// IImageDownloader download Image to disk
type IImageDownloader interface {
	Download(image *Image) (*File, error)
}
