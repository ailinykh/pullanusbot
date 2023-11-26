package api

import (
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
	default:
		medias := []*core.Media{}
		for _, m := range media {
			tmf.l.Infof("Type: %s", m.Type)
			switch m.Type {
			case "video", "animated_gif":
				//TODO: Codec ??
				medias = append(medias, &core.Media{
					ResourceURL: m.VideoInfo.best().URL,
					URL:         url, Title: tweet.User.Name,
					Description: tweet.FullText,
					Duration:    int(m.VideoInfo.Duration / 1000),
					Type:        core.TVideo,
				})
			case "photo":
				medias = append(medias, &core.Media{
					ResourceURL: m.MediaUrlHttps,
					URL:         url, Title: tweet.User.Name,
					Description: tweet.FullText,
					Type:        core.TPhoto,
				})
			default:
				tmf.l.Errorf("unexpected type: %s", m.Type)
			}
		}
		return medias, nil
	}
}
