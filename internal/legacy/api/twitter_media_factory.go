package api

import (
	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateTwitterMediaFactory(l core.Logger, t legacy.ITask) *TwitterMediaFactory {
	return &TwitterMediaFactory{l, CreateTwitterAPI(l, t)}
}

type TwitterMediaFactory struct {
	l   core.Logger
	api *TwitterAPI
}

// CreateMedia is a core.IMediaFactory interface implementation
func (tmf *TwitterMediaFactory) CreateMedia(tweetID string) ([]*legacy.Media, error) {
	tweet, err := tmf.api.getTweetByID(tweetID)
	if err != nil {
		tmf.l.Error(err)
		return nil, err
	}

	url := "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.ID
	media := tweet.ExtendedEntities.Media

	if len(media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		media = tweet.QuotedStatus.ExtendedEntities.Media
		tmf.l.Warn("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	switch len(media) {
	case 0:
		screenshot, err := tmf.api.getScreenshot(tweet)
		if err != nil {
			tmf.l.Error(err)
			return []*legacy.Media{{URL: url, Title: tweet.User.Name, Description: tweet.FullText, Type: legacy.TText}}, nil
		}
		return []*legacy.Media{{ResourceURL: screenshot.URL, URL: url, Title: tweet.User.Name, Description: "", Type: legacy.TPhoto}}, nil
	default:
		medias := []*legacy.Media{}
		for _, m := range media {
			tmf.l.Info("Type: %s", m.Type)
			switch m.Type {
			case "video", "animated_gif":
				//TODO: Codec ??
				medias = append(medias, &legacy.Media{
					ResourceURL: m.VideoInfo.best().URL,
					URL:         url, Title: tweet.User.Name,
					Description: tweet.FullText,
					Duration:    int(m.VideoInfo.Duration / 1000),
					Type:        legacy.TVideo,
				})
			case "photo":
				medias = append(medias, &legacy.Media{
					ResourceURL: m.MediaUrlHttps,
					URL:         url, Title: tweet.User.Name,
					Description: tweet.FullText,
					Type:        legacy.TPhoto,
				})
			default:
				tmf.l.Error("unexpected type: %s", m.Type)
			}
		}
		return medias, nil
	}
}
