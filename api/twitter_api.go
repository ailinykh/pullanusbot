package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterAPI is a default Twitter factory
func CreateTwitterAPI(l core.ILogger, t core.ITask) *TwitterAPI {
	return &TwitterAPI{l, t, []string{
		"AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw",
		"AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
	}}
}

// Twitter API
type TwitterAPI struct {
	l      core.ILogger
	task   core.ITask
	tokens []string
}

func (api *TwitterAPI) getTweetByID(tweetID string) (*Tweet, error) {
	var tweet *Tweet
	var err = fmt.Errorf("tokens not set")
	for _, t := range api.tokens {
		tweet, err = api.getTweetByIdAndToken(tweetID, t)
		if err == nil || !strings.HasPrefix(err.Error(), "Rate limit exceeded") {
			return tweet, err
		}
	}
	return tweet, err
}

func (api *TwitterAPI) getTweetByIdAndToken(tweetID string, token string) (*Tweet, error) {
	client := http.DefaultClient
	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
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

	// os.WriteFile("tweet-"+tweetID+".json", body, 0644)

	if len(tweet.Errors) > 0 {
		if tweet.Errors[0].Code == 88 { // "Rate limit exceeded 88"
			return nil, fmt.Errorf("%s %s", tweet.Errors[0].Message, res.Header["X-Rate-Limit-Reset"][0])
		}
		api.l.Errorf("%s %s", tweet.Errors, token)
		return nil, fmt.Errorf(tweet.Errors[0].Message)
	}

	return &tweet, err
}

func (api *TwitterAPI) getScreenshot(tweet *Tweet) (*TweetScreenshot, error) {
	ch := make(chan []byte)
	task := TweetScreenshot{TweetId: tweet.ID, Username: tweet.User.ScreenName}
	data, err := json.Marshal(task)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	err = api.task.Perform(data, ch)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}

	api.l.Infof("retreiving screenshot for %s/%s", tweet.User.ScreenName, tweet.ID)

	select {
	case data := <-ch:
		api.l.Info(string(data))

		var screenshot TweetScreenshot
		err = json.Unmarshal(data, &screenshot)
		if err != nil {
			api.l.Error(err)
			return nil, err
		}

		return &screenshot, nil

	case <-time.After(1 * time.Minute):
		return nil, fmt.Errorf("screenshot timeout for %s/%s", tweet.User.ScreenName, tweet.ID)
	}
}
