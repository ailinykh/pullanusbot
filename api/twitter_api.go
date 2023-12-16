package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
		guestToken, err := api.getGuestToken("https://twitter.com/username/status/" + tweetID)
		if err != nil {
			return nil, err
		}

		api.l.Infof("guest token received: %s", guestToken)

		api.credentials = TwitterApiCredentials{
			bearer_token: api.credentials.bearer_token,
			guest_token:  guestToken,
		}

		return api.getTweetFromGraphQL(tweetID)
	}

	return tweet, err
}

func (api *TwitterAPI) getGuestToken(url string) (string, error) {
	api.l.Info("updating guest token")

	client := http.DefaultClient
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	r := regexp.MustCompile(`gt=(\d+);`)
	match := r.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return "", fmt.Errorf("failed to parse guest_token from %s", url)
	}
	return match[1], nil
}

func (api *TwitterAPI) getTweetFromGraphQL(tweetID string) (*Tweet, error) {
	data, _ := json.Marshal(GraphQLVariables{false, tweetID, false, false})
	variables := url.QueryEscape(string(data))

	features := map[string]bool{
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"longform_notetweets_consumption_enabled":                                 true,
		"longform_notetweets_inline_media_enabled":                                true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"responsive_web_enhance_cards_enabled":                                    false,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_home_pinned_timelines_enabled":                            true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"standardized_nudges_misinfo":                                             true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"verified_phone_label_enabled":                                            false,
		"view_counts_everywhere_api_enabled":                                      true,
	}
	data, err := json.Marshal(features)
	if err != nil {
		api.l.Error(err)
		return nil, err
	}
	featuresString := url.QueryEscape(string(data))
	url := fmt.Sprintf("https://api.twitter.com/graphql/5GOHgZe-8U2j5sVHQzEm9A/TweetResultByRestId?variables=%s&features=%s", variables, featuresString)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("authorization", "Bearer "+api.credentials.bearer_token)
	req.Header.Add("x-guest-token", api.credentials.guest_token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response GraphQLResponse
	body, _ := io.ReadAll(res.Body)

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
	body, _ := io.ReadAll(res.Body)

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
