package core

type IFileDownloader interface {
	Download(URL) (*File, error)
}

type IFileUploader interface {
	Upload(*File) (URL, error)
}

type IImageDownloader interface {
	Download(image *Image) (*File, error)
}
