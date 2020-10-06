package twitter

type twitterReponse struct {
	ID               string          `json:"id_str"`
	FullText         string          `json:"full_text"`
	Entities         twitterEntity   `json:"entities"`
	ExtendedEntities twitterEntity   `json:"extended_entities,omitempty"`
	User             twitterUser     `json:"user"`
	QuotedStatus     *twitterReponse `json:"quoted_status,omitempty"`
	Errors           []twitterError  `json:"errors,omitempty"`
}

type twitterUser struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type twitterEntity struct {
	Urls  []twitterURL   `json:"urls,omitempty"`
	Media []twitterMedia `json:"media"`
}

type twitterMedia struct {
	MediaURL      string           `json:"media_url"`
	MediaURLHTTPS string           `json:"media_url_https"`
	Type          string           `json:"type"`
	VideoInfo     twitterVideoInfo `json:"video_info,omitempty"`
}

type twitterURL struct {
	ExpandedURL string `json:"expanded_url"`
}

type twitterVideoInfo struct {
	Variants []twitterVideoInfoVariant `json:"variants"`
}

type twitterError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (info *twitterVideoInfo) best() twitterVideoInfoVariant {
	variant := info.Variants[0]
	for _, v := range info.Variants {
		if v.ContentType == "video/mp4" && v.Bitrate > variant.Bitrate {
			return v
		}
	}
	return variant
}

type twitterVideoInfoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
