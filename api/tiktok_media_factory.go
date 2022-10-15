package api

import (
	"fmt"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokMediaFactory(l core.ILogger, api ITikTokAPI) core.IMediaFactory {
	return &TikTokMediaFactory{l, api}
}

type TikTokMediaFactory struct {
	l   core.ILogger
	api ITikTokAPI
}

func (factory *TikTokMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected url %s", url)
	}

	item, err := factory.api.GetItem(parts[3], parts[5])
	if err != nil {
		return nil, err
	}

	title := item.Desc
	if len(title) == 0 {
		title = fmt.Sprintf("%s (@%s)", item.Author.Nickname, item.Author.UniqueId)
	}

	description := fmt.Sprintf("%s (@%s) has created a short video on TikTok with music %s.", item.Author.Nickname, item.Author.UniqueId, item.Music.Title)
	for _, s := range item.Stickers {
		description = description + " | " + s
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: item.VideoURL,
		Title:       title,
		Description: description,
	}

	return []*core.Media{media}, nil
}
