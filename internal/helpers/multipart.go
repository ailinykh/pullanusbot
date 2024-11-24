package helpers

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func MultipartFrom(params map[string]io.Reader, writer *multipart.Writer) error {
	for key, reader := range params {
		var part io.Writer
		var err error

		switch r := reader.(type) {
		case *os.File:
			baseName := filepath.Base(r.Name())
			part, err = writer.CreateFormFile(key, baseName)
		default:
			part, err = writer.CreateFormField(key)
		}
		if err != nil {
			return err
		}

		_, err = io.Copy(part, reader)
		if err != nil {
			return err
		}
	}

	return nil
}
