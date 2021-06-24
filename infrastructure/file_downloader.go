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
func (FileDownloader) Download(url core.URL, filepath string) (*core.File, error) {
	name := path.Base(filepath)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return &core.File{Name: name, Path: filepath}, err
}
