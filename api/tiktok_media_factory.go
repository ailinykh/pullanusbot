package api

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokMediaFactory(l core.ILogger, api YoutubeApi) core.IMediaFactory {
	return &TikTokMediaFactory{l, api}
}

type TikTokMediaFactory struct {
	l   core.ILogger
	api YoutubeApi
}

func (factory *TikTokMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	item, err := factory.api.get(url)
	if err != nil {
		factory.l.Error(err)
		return nil, err
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: item.Url,
		Title:       fmt.Sprintf("%s (@%s)", item.Creator, item.Uploader),
		Description: fmt.Sprintf(`<a href="https://tiktok.com/@%s">(@%s)'s</a> short video with %s - %s.`, item.Uploader, item.Creator, item.Track, item.Artist),
	}

	return []*core.Media{media}, nil
}
