package image_uploader

import (
	"net/url"
	"os"
)

type Uploader interface {
	Upload(*os.File) (*url.URL, error)
}
