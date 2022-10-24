package infrastructure

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateFileDownloader is a default FileDownloader factory
func CreateFileDownloader(l core.ILogger) *FileDownloader {
	return &FileDownloader{l}
}

// FileDownloader is a default implementation for core.IFileDownloader
type FileDownloader struct {
	l core.ILogger
}

// Download is a core.IFileDownloader interface implementation
func (downloader *FileDownloader) Download(url core.URL, filepath string) (*core.File, error) {
	name := path.Base(filepath)
	downloader.l.Infof("downloading %s %s", url, filepath)
	// Get the data
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:59.0) Gecko/20100101 Firefox/59.0")
	req.Header.Set("Referer", url)
	res, err := client.Do(req)
	if err != nil {
		downloader.l.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		downloader.l.Error(err)
		return nil, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, res.Body)
	if err != nil {
		downloader.l.Error(err)
		return nil, err
	}

	// Retreive file size
	stat, err := os.Stat(filepath)
	if err != nil {
		downloader.l.Error(err)
		return nil, err
	}
	return &core.File{Name: name, Path: filepath, Size: stat.Size()}, err
}
