package api

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterMediaFactory(l core.ILogger, t core.ITask) *TwitterMediaFactory {
	return &TwitterMediaFactory{l, CreateTwitterAPI(l, t)}
}

type TwitterMediaFactory struct {
	l   core.ILogger
	api *TwitterAPI
}

// CreateMedia is a core.IMediaFactory interface implementation
func (tmf *TwitterMediaFactory) CreateMedia(tweetID string) ([]*core.Media, error) {
	tweet, err := tmf.api.getTweetByID(tweetID)
	if err != nil {
		tmf.l.Error(err)
		return nil, err
	}

	url := "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.ID
	media := tweet.ExtendedEntities.Media

	if len(media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		media = tweet.QuotedStatus.ExtendedEntities.Media
		tmf.l.Warningf("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	switch len(media) {
	case 0:
		screenshot, err := tmf.api.getScreenshot(tweet)
		if err != nil {
			tmf.l.Error(err)
			return []*core.Media{{URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: core.TText}}, nil
		}
		return []*core.Media{{ResourceURL: screenshot.URL, URL: url, Title: tweet.User.Name, Description: "", Type: core.TPhoto}}, nil
	case 1:
		if media[0].Type == "video" || media[0].Type == "animated_gif" {
			//TODO: Codec ??
			return []*core.Media{{
				ResourceURL: media[0].VideoInfo.best().URL,
				URL:         url, Title: tweet.User.Name,
				Description: tweet.FullText,
				Type:        core.TVideo,
			}}, nil
		} else if media[0].Type == "photo" {
			return []*core.Media{{
				ResourceURL: media[0].MediaUrlHttps,
				URL:         url, Title: tweet.User.Name,
				Description: tweet.FullText,
				Type:        core.TPhoto,
			}}, nil
		} else {
			return nil, fmt.Errorf("unexpected type: %s", media[0].Type)
		}
	default:
		// t.sendAlbum(media, tweet, m)
		medias := []*core.Media{}
		for _, m := range media {
			medias = append(medias, &core.Media{
				ResourceURL: m.MediaUrlHttps,
				URL:         url,
				Title:       tweet.User.Name,
				Description: tweet.FullText,
				Type:        core.TPhoto,
			})
		}
		return medias, nil
	}
}
