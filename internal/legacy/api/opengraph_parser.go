package api

import (
	"fmt"
	"html"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateOpenGraphParser(l core.Logger) *OpenGraphParser {
	return &OpenGraphParser{l}
}

type OpenGraphParser struct {
	l core.Logger
}

// CreateMedia is a core.IMediaFactory interface implementation
func (ogp *OpenGraphParser) CreateMedia(HTMLString string) ([]*legacy.Media, error) {
	video := ogp.parseMeta(HTMLString, "og:video")
	if len(video) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	video = html.UnescapeString(video)
	title := ogp.parseMeta(HTMLString, "og:title")
	description := ogp.parseMeta(HTMLString, "og:description")
	url := ogp.parseMeta(HTMLString, "og:url")

	media := &legacy.Media{
		ResourceURL: video,
		URL:         url,
		Title:       title,
		Description: description,
		Type:        legacy.TVideo,
	}
	return []*legacy.Media{media}, nil
}

func (ogp *OpenGraphParser) parseMeta(html string, property string) string {
	r := regexp.MustCompile(fmt.Sprintf(`<meta\s+property="%s"\s+content="([^"]+)"\/>`, property))
	match := r.FindStringSubmatch(html)
	if len(match) == 0 {
		ogp.l.Error("failed to parse meta content from %s", property)
		return ""
	}

	return match[1]
}
