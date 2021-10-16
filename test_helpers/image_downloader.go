package test_helpers

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateImageDownloader() *FakeImageDownloader {
	return &FakeImageDownloader{[]string{}, nil}
}

type FakeImageDownloader struct {
	Downloaded []string
	Err        error
}

// Upload is a core.IImageDownloader interface implementation
func (fid *FakeImageDownloader) Download(image *core.Image) (*core.File, error) {
	fid.Downloaded = append(fid.Downloaded, image.FileURL)
	return &core.File{Path: image.Path}, fid.Err
}
