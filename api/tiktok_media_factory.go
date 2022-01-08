package api

import (
	"fmt"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

type ITikTokAPI interface {
	GetItem(string, string) (*TikTokItemStruct, error)
}

func CreateTikTokMediaFactory(l core.ILogger, hc core.IHttpClient, r core.IRand) core.IMediaFactory {
	return &TikTokMediaFactory{l, &TikTokAPI{l, hc, r}}
}

type TikTokMediaFactory struct {
	l   core.ILogger
	api *TikTokAPI
}

func (factory *TikTokMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected url %s", url)
	}

	item, err := factory.api.getItemUsingJsonApi(parts[3], parts[5])
	if err != nil {
		item, err = factory.api.getItemUsingHtmlApi(parts[3], parts[5])

		if err != nil {
			return nil, err
		}
	}

	title := item.Desc
	if len(title) == 0 {
		title = fmt.Sprintf("%s (@%s)", item.Author.Nickname, item.Author.UniqueId)
	}

	description := fmt.Sprintf("%s (@%s) has created a short video on TikTok with music %s.", item.Author.Nickname, item.Author.UniqueId, item.Music.Title)
	for _, s := range item.StickersOnItem {
		for _, t := range s.StickerText {
			description = description + " | " + t
		}
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: item.Video.DownloadAddr,
		Title:       title,
		Description: description,
	}

	return []*core.Media{media}, nil
}
