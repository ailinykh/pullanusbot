package api

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterMediaFactory() *TwitterMediaFactory {
	return &TwitterMediaFactory{CreateTwitterAPI()}
}

type TwitterMediaFactory struct {
	api *TwitterAPI
}

// CreateMedia is a core.IMediaFactory interface implementation
func (tmf *TwitterMediaFactory) CreateMedia(tweetID string, author *core.User) ([]*core.Media, error) {
	tweet, err := tmf.api.getTweetByID(tweetID)
	if err != nil {
		return nil, err
	}

	if len(tweet.ExtendedEntities.Media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		tweet = tweet.QuotedStatus
		// logger.Warningf("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	media := tweet.ExtendedEntities.Media

	switch len(media) {
	case 0:
		return []*core.Media{{URL: "", Caption: tmf.makeCaption(author.Username, tweet), Type: core.TText}}, nil
	case 1:
		if media[0].Type == "video" || media[0].Type == "animated_gif" {
			//TODO: Codec ??
			return []*core.Media{{URL: media[0].VideoInfo.best().URL, Caption: tmf.makeCaption(author.Username, tweet), Type: core.TVideo}}, nil
		} else if media[0].Type == "photo" {
			return []*core.Media{{URL: media[0].MediaURL, Caption: tmf.makeCaption(author.Username, tweet), Type: core.TPhoto}}, nil
		} else {
			return nil, errors.New("unexpected type: " + media[0].Type)
		}
	default:
		// t.sendAlbum(media, tweet, m)
		medias := []*core.Media{}
		for _, m := range media {
			medias = append(medias, &core.Media{URL: m.MediaURL, Caption: tmf.makeCaption(author.Username, tweet), Type: core.TPhoto})
		}
		return medias, nil
	}
}

func (TwitterMediaFactory) makeCaption(author string, tweet *Tweet) string {
	re := regexp.MustCompile(`\s?http\S+$`)
	text := re.ReplaceAllString(tweet.FullText, "")
	return fmt.Sprintf("<a href='https://twitter.com/%s/status/%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", tweet.User.ScreenName, tweet.ID, tweet.User.Name, author, text)
}
