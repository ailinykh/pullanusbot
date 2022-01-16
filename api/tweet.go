package api

// Tweet is a twitter api representation of a single tweet
type Tweet struct {
	ID               string  `json:"id_str"`
	FullText         string  `json:"full_text"`
	Entities         Entity  `json:"entities"`
	ExtendedEntities Entity  `json:"extended_entities,omitempty"`
	User             User    `json:"user"`
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
	MediaURL      string    `json:"media_url"`
	MediaURLHTTPS string    `json:"media_url_https"`
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
