package api

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokMediaFactory(l core.ILogger, hc core.IHttpClient, r core.IRand) core.IMediaFactory {
	htmlApi := CreateTikTokHTMLAPI(l, hc, r)
	jsonApi := CreateTikTokJsonAPI(l, hc, r)
	return &TikTokMediaFactory{l, htmlApi, jsonApi}
}

type TikTokMediaFactory struct {
	l       core.ILogger
	htmlApi core.IMediaFactory
	jsonApi core.IMediaFactory
}

func (api *TikTokMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	media, err := api.jsonApi.CreateMedia(url)
	if err != nil {
		return api.htmlApi.CreateMedia(url)
	}

	return media, nil
}
