package api

import (
	"errors"
	"fmt"
	"html"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateOpenGraphParser(l core.ILogger) *OpenGraphParser {
	return &OpenGraphParser{l}
}

type OpenGraphParser struct {
	l core.ILogger
}

// CreateMedia is a core.IMediaFactory interface implementation
func (ogp *OpenGraphParser) CreateMedia(HTMLString string, author *core.User) ([]*core.Media, error) {
	video := ogp.parseMeta(HTMLString, "og:video")
	if len(video) == 0 {
		return nil, errors.New("video not found")
	}

	video = html.UnescapeString(video)
	title := ogp.parseMeta(HTMLString, "og:title")
	description := ogp.parseMeta(HTMLString, "og:description")
	url := ogp.parseMeta(HTMLString, "og:url")

	media := &core.Media{
		URL:     video,
		Type:    core.TVideo,
		Caption: fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, title, author.Username, description),
	}
	return []*core.Media{media}, nil
}

func (ogp *OpenGraphParser) parseMeta(html string, property string) string {
	r := regexp.MustCompile(fmt.Sprintf(`<meta\s+property="%s"\s+content="([^"]+)"\/>`, property))
	match := r.FindStringSubmatch(html)
	if len(match) == 0 {
		ogp.l.Errorf("can't find %s", property)
		return ""
	}

	return match[1]
}
