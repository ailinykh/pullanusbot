package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateTwitterAPI is a default Twitter factory
func CreateTwitterAPI(l core.ILogger, t core.ITask) *TwitterAPI {
	return &TwitterAPI{l, t, TwitterApiCredentials{
		bearer_token: "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		guest_token:  "1679397394880888834",
	},
	}
}

type TwitterApiCredentials struct {
	bearer_token string
	guest_token  string
}

// Twitter API
type TwitterAPI struct {
	l           core.ILogger
	task        core.ITask
	credentials TwitterApiCredentials
}

func (api *TwitterAPI) getTweetByID(tweetID string) (*Tweet, error) {
	tweet, err := api.getTweetFromGraphQL(tweetID)
	if err == nil {
		return tweet, err
	}

	if err.Error() == "Bad guest token" {
		resp, err := api.getGuestToken()
		if err != nil {
			return nil, err
		}

		api.l.Infof("guest token received %s", resp.GuestToken)

		api.credentials = TwitterApiCredentials{
			bearer_token: api.credentials.bearer_token,
			guest_token:  resp.GuestToken,
		}

		return api.getTweetFromGraphQL(tweetID)
	}

	return tweet, err
}

func (api *TwitterAPI) getGuestToken() (*GuestTokenResponse, error) {
	api.l.Info("updating guest token")

	req, _ := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	req.Header.Add("Authorization", "Bearer "+api.credentials.bearer_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response GuestTokenResponse
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (api *TwitterAPI) getTweetFromGraphQL(tweetID string) (*Tweet, error) {
	data, _ := json.Marshal(GraphQLVariables{tweetID, false, false, false})
	variables := url.QueryEscape(string(data))

	data, _ = json.Marshal(GraphQLFeatures{
		CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
		FreedomOfSpeechNotReachFetceEnabled:                            true,
		GraphqlIsTranslatableRwebTweetIsTranslatableEnabled:            true,
		LongformNotetweetsConsumptionEnabled:                           true,
		LongformNotetweetsInlineMediaEnabled:                           true,
		LongformNotetweetsRichTextReadEnabled:                          true,
		ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      false,
		ResponsiveWebEditTweetApiEnabled:                               true,
		ResponsiveWebEnhanceCardsEnabled:                               false,
		ResponsiveWebMediaDownloadVideoEnabled:                         true,
		ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
		ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
		ResponsiveWebTwitterArticleTweetConsumptionEnabled:             false,
		StandardizedNudgesMisinfo:                                      true,
		TweetAwardsWebTippingEnabled:                                   false,
		TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: true,
		TweetypieUnmentionOptimizationEnabled:                          true,
		VerifiedPhoneLabelEnabled:                                      false,
		ViewCountsEverywhereApiEnabled:                                 true,
	})
	features := url.QueryEscape(string(data))

	data, _ = json.Marshal(GraphQLFieldToggles{false})
	field_toggles := url.QueryEscape(string(data))

	url := fmt.Sprintf("https://twitter.com/i/api/graphql/2ICDjqPd81tulZcYrtpTuQ/TweetResultByRestId?variables=%s&features=%s&fieldToggles=%s", variables, features, field_toggles)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("authorization", "Bearer "+api.credentials.bearer_token)
	req.Header.Add("x-guest-token", api.credentials.guest_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response GraphQLResponse
	body, _ := ioutil.ReadAll(res.Body)

	// os.WriteFile("tweet-"+tweetID+".json", body, 0644)

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Errors) > 0 {
		if response.Errors[0].Code == 88 { // "Rate limit exceeded 88"
			return nil, fmt.Errorf("%s %s", response.Errors[0].Message, res.Header["X-Rate-Limit-Reset"][0])
		}
		return nil, fmt.Errorf(response.Errors[0].Message)
	}

	// TODO: combine `twitter_api` with `twitter_media_factory`
	tweet := response.Data.TweetResult.Result.Legacy
	user := response.Data.TweetResult.Result.Core.UserResults.Result.Legacy
	tweet.User = User{Name: user.Name, ScreenName: user.ScreenName}

	return &tweet, nil
}

func (api *TwitterAPI) getTweetByIdAndToken(tweetID string, creds TwitterApiCredentials) (*Tweet, error) {
	client := http.DefaultClient
	url := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+creds.bearer_token)
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
		api.l.Errorf("%s %s", tweet.Errors, creds.guest_token)
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
