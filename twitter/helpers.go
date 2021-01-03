package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

type (
	// IHelper makes testing easy
	IHelper interface {
		getTweet(string) (Tweet, error)
		makeAlbum([]Media, string) tb.Album
		makeCaption(*tb.Message, Tweet) string
	}

	// Helper is a default IHelper implementation
	Helper struct {
	}
)

func (Helper) getTweet(tweetID string) (Tweet, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID), nil)
	// Не ссыкотно в открытом репозитории хранить?
	req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw")
	res, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return Tweet{}, err
	}
	defer res.Body.Close()

	var tweet Tweet
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(body, &tweet)
	if err != nil {
		logger.Error(err)
		return Tweet{}, err
	}

	tweet.Header = res.Header
	return tweet, err
}

func (Helper) makeAlbum(media []Media, caption string) tb.Album {
	var photo *tb.Photo
	var album = tb.Album{}

	for i, m := range media {
		photo = &tb.Photo{File: tb.FromURL(m.MediaURL)}
		if i == len(media)-1 {
			photo.Caption = caption
			photo.ParseMode = tb.ModeHTML
		}
		album = append(album, photo)
	}

	return album
}

func (Helper) makeCaption(m *tb.Message, tweet Tweet) string {
	re := regexp.MustCompile(`\s?http\S+$`)
	text := re.ReplaceAllString(tweet.FullText, "")
	return fmt.Sprintf("<a href='https://twitter.com/%s/status/%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", tweet.User.ScreenName, tweet.ID, tweet.User.Name, m.Sender.Username, text)
}
