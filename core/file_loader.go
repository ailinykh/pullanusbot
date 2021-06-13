package core

type IFileDownloader interface {
	Download(URL) (*File, error)
}

type IFileUploader interface {
	Upload(*File) (URL, error)
}