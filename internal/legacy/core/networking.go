package core

import "os"

// IFileDownloader turns URL to File
type IFileDownloader interface {
	Download(URL, string) (*File, error)
}

// IFileUploader turns File to URL
type IFileUploader interface {
	Upload(*File) (URL, error)
}

// IImageDownloader download Image to disk
type IImageDownloader interface {
	Download(image *Image) (*os.File, error)
}

// IHttpClient retreives remote content info
type IHttpClient interface {
	GetContentType(URL) (string, error)
	GetContent(URL) (string, error)
	GetRedirectLocation(url URL) (URL, error)
	SetHeader(string, string)
}
