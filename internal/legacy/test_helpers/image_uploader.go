package test_helpers

import (
	"net/url"
	"os"
)

func CreateImageUploader() *FakeImageUploader {
	return &FakeImageUploader{[]string{}, nil}
}

type FakeImageUploader struct {
	Uploaded []string
	Err      error
}

// Upload is a core.IImageUploader interface implementation
func (ffu *FakeImageUploader) Upload(file *os.File) (*url.URL, error) {
	ffu.Uploaded = append(ffu.Uploaded, file.Name())
	return &url.URL{}, ffu.Err
}
