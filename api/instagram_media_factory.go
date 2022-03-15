package api

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateInstagramMediaFactory(l core.ILogger, cookiesFile string) *InstagramMediaFactory {
	return &InstagramMediaFactory{l, CreateInstagramAPI(l, cookiesFile)}
}

type InstagramMediaFactory struct {
	l   core.ILogger
	api *InstagramAPI
}

// CreateMedia is a core.IMediaFactory interface implementation
func (factory *InstagramMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	reel, err := factory.api.GetReel(url)
	if err != nil {
		factory.l.Error(err)
		return nil, err
	}

	if len(reel.Items) < 1 {
		return nil, fmt.Errorf("insufficient reel items")
	}

	item := reel.Items[0]
	return []*core.Media{{ResourceURL: item.VideoVersions[0].URL, URL: "https://www.instagram.com/reel/" + item.Code + "/", Title: item.User.FullName, Caption: item.Caption.Text}}, nil
}
