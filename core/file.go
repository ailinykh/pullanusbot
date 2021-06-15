package core

import "os"

type File struct {
	Name string
	Path string
}

func (f *File) Dispose() error {
	return os.Remove(f.Path)
}
