package core

type IFileDownloader interface {
	Download(string, string) error
}
