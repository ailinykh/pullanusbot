package core

import (
	"os"
	"path"
	"sync"
)

func CreateImage(id string, dl func() (string, error)) Image {
	return Image{ID: id, dl: dl}
}

type Image struct {
	File
	ID string

	dl    func() (string, error)
	mutex sync.Mutex
}

func (i *Image) Download() error {
	if _, err := os.Stat(i.File.Path); err == nil {
		return nil
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	filepath, err := i.dl()
	if err != nil {
		return nil
	}

	i.File.Name = path.Base(filepath)
	i.File.Path = filepath

	return nil
}

func (i *Image) Dispose() error {
	return os.Remove(i.File.Path)
}
