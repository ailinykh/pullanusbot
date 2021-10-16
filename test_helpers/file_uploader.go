package test_helpers

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateFileUploader() *FakeFileUploader {
	return &FakeFileUploader{[]string{}, nil}
}

type FakeFileUploader struct {
	Uploaded []string
	Err      error
}

// Upload is a core.IFileUploader interface implementation
func (ffu *FakeFileUploader) Upload(file *core.File) (core.URL, error) {
	ffu.Uploaded = append(ffu.Uploaded, file.Path)
	return file.Path, ffu.Err
}
