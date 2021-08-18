package api

import (
	"errors"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterMediaFactory(l core.ILogger) *TwitterMediaFactory {
	return &TwitterMediaFactory{l, CreateTwitterAPI()}
}

type TwitterMediaFactory struct {
	l   core.ILogger
	api *TwitterAPI
}

// CreateMedia is a core.IMediaFactory interface implementation
func (tmf *TwitterMediaFactory) CreateMedia(tweetID string, _ *core.User) ([]*core.Media, error) {
	tweet, err := tmf.api.getTweetByID(tweetID)
	if err != nil {
		return nil, err
	}

	if len(tweet.ExtendedEntities.Media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		tweet = tweet.QuotedStatus
		tmf.l.Warningf("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	media := tweet.ExtendedEntities.Media
	url := "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.ID

	switch len(media) {
	case 0:
		return []*core.Media{{URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: core.TText}}, nil
	case 1:
		if media[0].Type == "video" || media[0].Type == "animated_gif" {
			//TODO: Codec ??
			return []*core.Media{{ResourceURL: media[0].VideoInfo.best().URL, URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: core.TVideo}}, nil
		} else if media[0].Type == "photo" {
			return []*core.Media{{ResourceURL: media[0].MediaURL, URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: core.TPhoto}}, nil
		} else {
			return nil, errors.New("unexpected type: " + media[0].Type)
		}
	default:
		// t.sendAlbum(media, tweet, m)
		medias := []*core.Media{}
		for _, m := range media {
			medias = append(medias, &core.Media{ResourceURL: m.MediaURL, URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: core.TPhoto})
		}
		return medias, nil
	}
}
