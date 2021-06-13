package infrastructure

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

type FileDownloader struct{}

// core.IFileDownloader
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
