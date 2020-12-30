package twitter

import "net/http"

// Tweet is a generic twitter response struct
type Tweet struct {
	ID               string  `json:"id_str"`
	FullText         string  `json:"full_text"`
	Entities         Entity  `json:"entities"`
	ExtendedEntities Entity  `json:"extended_entities,omitempty"`
	User             User    `json:"user"`
	QuotedStatus     *Tweet  `json:"quoted_status,omitempty"`
	Errors           []Error `json:"errors,omitempty"`
	Header           http.Header
}

// User represents tweet author
type User struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

// Entity is an attachment
type Entity struct {
	Urls  []URL   `json:"urls,omitempty"`
	Media []Media `json:"media"`
}

// Media represents picture, video of gif
type Media struct {
	MediaURL      string    `json:"media_url"`
	MediaURLHTTPS string    `json:"media_url_https"`
	Type          string    `json:"type"`
	VideoInfo     VideoInfo `json:"video_info,omitempty"`
}

// URL struct
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

// VideoInfoVariant represents different video formats
type VideoInfoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
