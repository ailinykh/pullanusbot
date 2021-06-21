package core

import "os"

// File ...
type File struct {
	Name string
	Path string
	Size int64
}

// Dispose for filesystem cleanup
func (f *File) Dispose() error {
	return os.Remove(f.Path)
}
