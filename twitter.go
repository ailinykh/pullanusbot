package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Twitter goes to prettify twitter links
type Twitter struct {
}

type twitterReponse struct {
	ID               string          `json:"id_str"`
	FullText         string          `json:"full_text"`
	ExtendedEntities twitterEntity   `json:"extended_entities"`
	User             twitterUser     `json:"user"`
	QuotedStatus     *twitterReponse `json:"quoted_status"`
}

type twitterUser struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type twitterEntity struct {
	Media []twitterMedia `json:"media"`
}

type twitterMedia struct {
	MediaURL  string           `json:"media_url"`
	Type      string           `json:"type"`
	VideoInfo twitterVideoInfo `json:"video_info"`
}

type twitterVideoInfo struct {
	Variants []twitterVideoInfoVariant `json:"variants"`
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

func (t *Twitter) initialize() {
	bot.Handle(tb.OnText, t.checkMessage)
	log.Println("Twitter: successfully initialized")
}

func (t *Twitter) checkMessage(m *tb.Message) {
	b, ok := bot.(*tb.Bot)
	if !ok {
		log.Println("Twitter: Bot cast failed")
		return
	}

	r, _ := regexp.Compile(`twitter\.com.+/(\d+)`)
	match := r.FindStringSubmatch(m.Text)
	if len(match) > 1 {
		tweetID := match[1]
		log.Printf("Twitter: Found tweet %s", tweetID)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID), nil)
		req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw")
		res, err := client.Do(req)
		if err != nil {
			log.Printf("Twitter: Tweet json fetch error: %s", err)
			return
		}
		defer res.Body.Close()

		var twResp twitterReponse
		body, _ := ioutil.ReadAll(res.Body)

		json.Unmarshal(body, &twResp)

		caption := t.getCaption(m, twResp)
		album := t.getAlbum(twResp.ExtendedEntities.Media)

		if len(twResp.ExtendedEntities.Media) == 0 && len(twResp.QuotedStatus.ExtendedEntities.Media) > 0 {
			caption = t.getCaption(m, *twResp.QuotedStatus)
			album = t.getAlbum(twResp.QuotedStatus.ExtendedEntities.Media)
		}

		switch len(album) {
		case 0:
			_, err = b.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeMarkdown, DisableWebPagePreview: true})
		case 1:
			f, ok := album[0].(*tb.Video)
			if ok {
				f.Caption = caption
				_, err = f.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			} else {
				f, ok := album[0].(*tb.Photo)
				if ok {
					f.Caption = caption
					_, err = f.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				}
			}
		default:
			b.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeMarkdown, DisableWebPagePreview: true})
			_, err = b.SendAlbum(m.Chat, album)
		}

		if err == nil {
			log.Println("Twitter: Messages sent. Deleting original")
			err = b.Delete(m)
			if err != nil {
				log.Printf("Twitter: Can't delete original message: %s", err)
			}
		} else {
			log.Printf("Twitter: Can't send video: %s", err)
		}
	}
}

func (t *Twitter) getAlbum(media []twitterMedia) tb.Album {
	var file tb.Sendable
	var album = tb.Album{}

	for _, m := range media {
		if m.Type == "video" {
			file = &tb.Video{File: tb.FromURL(m.VideoInfo.best().URL)}
		} else if m.Type == "photo" {
			file = &tb.Photo{File: tb.FromURL(m.MediaURL)}
		} else {
			log.Printf("Twitter: Unknown type: %s", m.Type)
			file = nil
		}

		f, ok := file.(tb.InputMedia)
		if ok {
			album = append(album, f)
		}
	}

	return album
}

func (t *Twitter) getCaption(m *tb.Message, r twitterReponse) string {
	return fmt.Sprintf("[🐦](https://twitter.com/%s/status/%s) *%s* _(by %s)_\n%s", r.User.ScreenName, r.ID, r.User.Name, m.Sender.Username, r.FullText)
}
