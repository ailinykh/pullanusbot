package api

// Tweet is a twitter api representation of a single tweet
type Tweet struct {
	ID               string  `json:"id_str"`
	FullText         string  `json:"full_text"`
	Entities         Entity  `json:"entities"`
	ExtendedEntities Entity  `json:"extended_entities,omitempty"`
	User             User    `json:"user,omitempty"`
	QuotedStatus     *Tweet  `json:"quoted_status,omitempty"`
	Errors           []Error `json:"errors,omitempty"`
}

// User ...
type User struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

// Entity ...
type Entity struct {
	Urls  []URL   `json:"urls,omitempty"`
	Media []Media `json:"media"`
}

// Media ...
type Media struct {
	MediaUrlHttps string    `json:"media_url_https"`
	Type          string    `json:"type"`
	VideoInfo     VideoInfo `json:"video_info,omitempty"`
}

// URL ...
type URL struct {
	ExpandedURL string `json:"expanded_url"`
}

// VideoInfo ...
type VideoInfo struct {
	Variants []VideoInfoVariant `json:"variants"`
}

// Error ...
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (info *VideoInfo) best() VideoInfoVariant {
	variant := info.Variants[0]
	for _, v := range info.Variants {
		if v.ContentType == "video/mp4" && v.Bitrate > variant.Bitrate {
			return v
		}
	}
	return variant
}

// VideoInfoVariant ...
type VideoInfoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

type TweetScreenshot struct {
	TweetId  string `json:"tweetId"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

// GraphQL types
type GraphQLRequest struct {
	Variables    GraphQLVariables    `json:"variables"`
	Features     GraphQLFeatures     `json:"features"`
	FieldToggles GraphQLFieldToggles `json:"fieldToggles"`
}

type GraphQLVariables struct {
	TweetId                string `json:"tweetId"`
	WithCommunity          bool   `json:"withCommunity"`
	IncludePromotedContent bool   `json:"includePromotedContent"`
	WithVoice              bool   `json:"withVoice"`
}

type GraphQLFeatures struct {
	CreatorSubscriptionsTweetPreviewApiEnabled                     bool `json:"creator_subscriptions_tweet_preview_api_enabled"`
	FreedomOfSpeechNotReachFetceEnabled                            bool `json:"freedom_of_speech_not_reach_fetch_enabled"`
	GraphqlIsTranslatableRwebTweetIsTranslatableEnabled            bool `json:"graphql_is_translatable_rweb_tweet_is_translatable_enabled"`
	LongformNotetweetsConsumptionEnabled                           bool `json:"longform_notetweets_consumption_enabled"`
	LongformNotetweetsInlineMediaEnabled                           bool `json:"longform_notetweets_inline_media_enabled"`
	LongformNotetweetsRichTextReadEnabled                          bool `json:"longform_notetweets_rich_text_read_enabled"`
	ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled      bool `json:"responsive_web_graphql_skip_user_profile_image_extensions_enabled"`
	ResponsiveWebEditTweetApiEnabled                               bool `json:"responsive_web_edit_tweet_api_enabled"`
	ResponsiveWebEnhanceCardsEnabled                               bool `json:"responsive_web_enhance_cards_enabled"`
	ResponsiveWebMediaDownloadVideoEnabled                         bool `json:"responsive_web_media_download_video_enabled"`
	ResponsiveWebGraphqlTimelineNavigationEnabled                  bool `json:"responsive_web_graphql_timeline_navigation_enabled"`
	ResponsiveWebGraphqlExcludeDirectiveEnabled                    bool `json:"responsive_web_graphql_exclude_directive_enabled"`
	ResponsiveWebTwitterArticleTweetConsumptionEnabled             bool `json:"responsive_web_twitter_article_tweet_consumption_enabled"`
	StandardizedNudgesMisinfo                                      bool `json:"standardized_nudges_misinfo"`
	TweetAwardsWebTippingEnabled                                   bool `json:"tweet_awards_web_tipping_enabled"`
	TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled bool `json:"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled"`
	TweetypieUnmentionOptimizationEnabled                          bool `json:"tweetypie_unmention_optimization_enabled"`
	VerifiedPhoneLabelEnabled                                      bool `json:"verified_phone_label_enabled"`
	ViewCountsEverywhereApiEnabled                                 bool `json:"view_counts_everywhere_api_enabled"`
}

type GraphQLFieldToggles struct {
	WithArticleRichContentState bool `json:"withArticleRichContentState"`
}

type GraphQLResponse struct {
	Errors []Error             `json:"errors,omitempty"`
	Data   GraphQLResponseData `json:"data"`
}

type GraphQLResponseData struct {
	TweetResult GraphQLResponseTweetResult `json:"tweetResult"`
}

type GraphQLResponseTweetResult struct {
	Result GraphQLResponseTweetResultResult `json:"result"`
}

type GraphQLResponseTweetResultResult struct {
	Core   GraphQLResponseCore `json:"core"`
	Legacy Tweet               `json:"legacy"`
	RestId string              `json:"rest_id"`
}

type GraphQLResponseCore struct {
	UserResults GraphQLResponseUserResults `json:"user_results"`
}

type GraphQLResponseUserResults struct {
	Result GraphQLResponseUserResult `json:"result"`
}

type GraphQLResponseUserResult struct {
	Legacy   User   `json:"legacy"`
	RestId   string `json:"rest_id"`
	Verified bool   `json:"is_blue_verified"`
}

type GuestTokenResponse struct {
	GuestToken string `json:"guest_token"`
}
