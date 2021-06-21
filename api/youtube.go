package api

import (
	"errors"
	"strings"
)

// Video is a struct to handle youtube-dl's JSON output
type Video struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Duration    int          `json:"duration"`
	Formats     []*Format    `json:"formats"`
	Title       string       `json:"title"`
	Thumbnail   string       `json:"thumbnail"` // might be .webp
	Thumbnails  []*Thumbnail `json:"thumbnails"`
}

func (v Video) audioFormat() (*Format, error) {
	for _, f := range v.Formats {
		if f.FormatID == "140" {
			return f, nil
		}
	}

	return nil, errors.New("140 not found for " + v.ID)
}

// might be empty
func (v Video) availableFormats() []*Format {
	rv := []*Format{}
	for _, f := range v.Formats {
		if f.Ext == "mp4" { // webm not friendly for iPhone
			if f.VCodec != "none" && f.ACodec == "none" {
				if strings.HasSuffix(f.FormatNote, "p") || strings.Contains(f.FormatNote, "DASH") { // skip 720p60
					rv = append(rv, f)
				}
			}
		}
	}
	return rv
}

func (v Video) formatByID(id string) (*Format, error) {
	for _, f := range v.Formats {
		if f.FormatID == id {
			return f, nil
		}
	}
	return nil, errors.New("can't find format with id " + id)
}

func (v Video) thumb() *Thumbnail {
	th := v.Thumbnails[0]
	for _, t := range v.Thumbnails {
		if !strings.Contains(t.URL, ".webp") {
			th = t
		}
	}
	return th
}

// Format is a description of available formats for downloading
type Format struct {
	Ext        string `json:"ext"`
	Filesize   int    `json:"filesize"`
	Format     string `json:"format"`
	FormatID   string `json:"format_id"`
	FormatNote string `json:"format_note"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	VCodec     string `json:"vcodec"`
	ACodec     string `json:"acodec"`
}

// Thumbnail is a low resolution picture
type Thumbnail struct {
	ID         string `json:"id"`
	Resolution string `json:"resolution"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	URL        string `json:"url"`
}
