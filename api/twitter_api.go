package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterAPI is a default Twitter factory
func CreateTwitterAPI() *TwitterAPI {
	return &TwitterAPI{}
}

// Twitter API
type TwitterAPI struct{}

func (TwitterAPI) get(tweetID string) (*Tweet, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID), nil)
	req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var tweet Tweet
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(body, &tweet)
	if err != nil {
		return nil, err
	}

	if len(tweet.Errors) > 0 {
		if tweet.Errors[0].Code == 88 { // "Rate limit exceeded 88"
			return nil, errors.New(tweet.Errors[0].Message + " " + res.Header["X-Rate-Limit-Reset"][0])
		}
		return nil, errors.New(tweet.Errors[0].Message)
	}

	return &tweet, err
}

// CreateMedia is a core.IMediaFactory interface implementation
func (t *TwitterAPI) CreateMedia(tweetID string, author *core.User) ([]*core.Media, error) {
	tweet, err := t.get(tweetID)
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
		return []*core.Media{{URL: "", Caption: t.makeCaption(author.Username, tweet), Type: core.TText}}, nil
	case 1:
		if media[0].Type == "video" || media[0].Type == "animated_gif" {
			return []*core.Media{{URL: media[0].VideoInfo.best().URL, Caption: t.makeCaption(author.Username, tweet), Type: core.TVideo}}, nil
		} else if media[0].Type == "photo" {
			return []*core.Media{{URL: media[0].MediaURL, Caption: t.makeCaption(author.Username, tweet), Type: core.TPhoto}}, nil
		} else {
			return nil, errors.New("Unknown type: " + media[0].Type)
		}
	default:
		// t.sendAlbum(media, tweet, m)
		medias := []*core.Media{}
		for _, m := range media {
			medias = append(medias, &core.Media{URL: m.MediaURL, Caption: t.makeCaption(author.Username, tweet), Type: core.TPhoto})
		}
		return medias, nil
	}
}

func (TwitterAPI) makeCaption(author string, tweet *Tweet) string {
	re := regexp.MustCompile(`\s?http\S+$`)
	text := re.ReplaceAllString(tweet.FullText, "")
	return fmt.Sprintf("<a href='https://twitter.com/%s/status/%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", tweet.User.ScreenName, tweet.ID, tweet.User.Name, author, text)
}
