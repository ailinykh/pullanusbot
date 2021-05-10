package infrastructure

import (
	"io"
	"net/http"
	"os"
)

func CreateFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

type FileDownloader struct{}

func (FileDownloader) Download(url string, path string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
