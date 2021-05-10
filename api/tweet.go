package api

type Tweet struct {
	ID               string  `json:"id_str"`
	FullText         string  `json:"full_text"`
	Entities         Entity  `json:"entities"`
	ExtendedEntities Entity  `json:"extended_entities,omitempty"`
	User             User    `json:"user"`
	QuotedStatus     *Tweet  `json:"quoted_status,omitempty"`
	Errors           []Error `json:"errors,omitempty"`
}

type User struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type Entity struct {
	Urls  []URL   `json:"urls,omitempty"`
	Media []Media `json:"media"`
}

type Media struct {
	MediaURL      string    `json:"media_url"`
	MediaURLHTTPS string    `json:"media_url_https"`
	Type          string    `json:"type"`
	VideoInfo     VideoInfo `json:"video_info,omitempty"`
}

type URL struct {
	ExpandedURL string `json:"expanded_url"`
}

type VideoInfo struct {
	Variants []VideoInfoVariant `json:"variants"`
}

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

type VideoInfoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
