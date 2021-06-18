package infrastructure

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateFileDownloader is a default FileDownloader factory
func CreateFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

// FileDownloader is a default implementation for core.IFileDownloader
type FileDownloader struct{}

// Download is a core.IFileDownloader interface implementation
func (FileDownloader) Download(url core.URL) (*core.File, error) {
	name := path.Base(url)
	path := path.Join(os.TempDir(), name)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return &core.File{Name: name, Path: path}, err
}
