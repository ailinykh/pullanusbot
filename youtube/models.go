package youtube

import (
	"strings"

	"github.com/google/logger"
)

// Video is a struct to handle youtube-dl's JSON output
type Video struct {
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Duration    int         `json:"duration"`
	Formats     []Format    `json:"formats"`
	Title       string      `json:"title"`
	Thumbnail   string      `json:"thumbnail"` // might be .webp
	Thumbnails  []Thumbnail `json:"thumbnails"`
}

func (v Video) audioFormat() Format {
	for _, f := range v.Formats {
		if f.FormatID == "140" {
			return f
		}
	}
	logger.Warning("140 not found for ", v.ID)
	return v.Formats[0]
}

func (v Video) availableFormats() []Format {
	rv := []Format{}
	for _, f := range v.Formats {
		if f.Ext == "mp4" { // webm not friendly for iPhone
			if f.VCodec != "none" && f.ACodec == "none" {
				if strings.HasSuffix(f.FormatNote, "p") || strings.Contains(f.FormatNote, "DASH") { // skip 720p60
					rv = append(rv, f)
				}
			}
		}
	}

	if len(rv) == 0 {
		logger.Warningf("no available formats found for %s", v.ID)
	}

	return rv
}

func (v Video) formatByID(id string) Format {
	for _, f := range v.Formats {
		if f.FormatID == id {
			return f
		}
	}
	logger.Errorf("format with id %s not found", id)
	return v.availableFormats()[0]
}

func (v Video) thumb() Thumbnail {
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
