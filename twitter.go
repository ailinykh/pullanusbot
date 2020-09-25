package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Twitter goes to prettify twitter links
type Twitter struct {
}

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

func (t *Twitter) handleTextMessage(m *tb.Message) {
	r, _ := regexp.Compile(`twitter\.com.+/(\d+)\S*$`)
	match := r.FindStringSubmatch(m.Text)
	if len(match) > 1 {
		t.processTweet(m, match[1])
	}
}

func (t *Twitter) processTweet(m *tb.Message, tweetID string) {
	b, ok := bot.(*tb.Bot)
	if !ok {
		logger.Errorf("Bot cast failed")
		return
	}

	logger.Infof("Processing tweet %s", tweetID)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID), nil)
	req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw")
	res, err := client.Do(req)
	if err != nil {
		logger.Errorf("json fetch error: %v", err)
		return
	}
	defer res.Body.Close()

	var twResp twitterReponse
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(body, &twResp)
	if err != nil {
		logger.Errorf("json parse error: %v", err)
		return
	}

	if len(twResp.Errors) > 0 {
		if twResp.Errors[0].Code == 88 { // "Rate limit exceeded 88"
			limit, err := strconv.ParseInt(res.Header["X-Rate-Limit-Reset"][0], 10, 64)

			if err != nil {
				logger.Error(err)
				return
			}

			timeout := limit - time.Now().Unix()
			logger.Infof("Twitter api timeout %d seconds", timeout)

			go func() {
				time.Sleep(time.Duration(timeout) * time.Second)
				t.processTweet(m, tweetID)
			}()

			return
		}

		logger.Errorf("Twitter api error: %v", twResp.Errors)
		logger.Errorf("%v", res.Header)

		b.Send(m.Chat, fmt.Sprint(twResp.Errors), &tb.SendOptions{ReplyTo: m})
		return
	}

	caption := t.getCaption(m, twResp)
	media := twResp.ExtendedEntities.Media

	if len(twResp.ExtendedEntities.Media) == 0 && twResp.QuotedStatus != nil && len(twResp.QuotedStatus.ExtendedEntities.Media) > 0 {
		caption = t.getCaption(m, *twResp.QuotedStatus)
		media = twResp.QuotedStatus.ExtendedEntities.Media
	}

	switch len(media) {
	case 0:
		logger.Info("Sending as text")

		for _, url := range twResp.Entities.Urls {
			caption += "\n" + url.ExpandedURL
		}

		_, err = b.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true})
	case 1:
		if media[0].Type == "video" {
			file := &tb.Video{File: tb.FromURL(media[0].VideoInfo.best().URL)}
			file.Caption = caption
			logger.Infof("Sending as Video %s", file.FileURL)
			b.Notify(m.Chat, tb.UploadingVideo)
			_, err = file.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
		} else if media[0].Type == "photo" {
			file := &tb.Photo{File: tb.FromURL(media[0].MediaURL)}
			file.Caption = caption
			logger.Infof("Sending as Photo %s", file.FileURL)
			b.Notify(m.Chat, tb.UploadingPhoto)
			_, err = file.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
		} else {
			logger.Infof("Unknown type: %s", media[0].Type)
			b.Send(m.Chat, fmt.Sprintf("Unknown type: %s", media[0].Type), &tb.SendOptions{ReplyTo: m})
			return
		}
	default:
		logger.Infof("Sending as Album")
		b.Notify(m.Chat, tb.UploadingPhoto)
		_, err = b.SendAlbum(m.Chat, t.getAlbum(media, twResp.FullText))
	}

	if err == nil {
		logger.Info("Messages sent. Deleting original")
		err = b.Delete(m)
		if err != nil {
			logger.Errorf("Can't delete original message: %v", err)
		}
	} else {
		logger.Error(err)

		if strings.HasSuffix(err.Error(), "failed to get HTTP URL content") || strings.HasSuffix(err.Error(), "wrong file identifier/HTTP URL specified") {
			// Try to upload file to telegram
			logger.Info("Sending by uploading")

			filename := path.Base(media[0].VideoInfo.best().URL)
			filepath := path.Join(os.TempDir(), filename)
			defer os.Remove(filepath)

			err = downloadFile(filepath, media[0].VideoInfo.best().URL)
			if err != nil {
				logger.Errorf("video download error: %v", err)
				return
			}

			videofile, err := NewVideoFile(filepath)
			if err != nil {
				logger.Errorf("Can't create video file for %s, %v", filepath, err)
				return
			}
			defer os.Remove(videofile.filepath)
			defer os.Remove(videofile.thumbpath)
			caption := fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, m.Text, filename, m.Sender.Username)
			uploadFile(videofile, m, caption)
		}
	}
}

func (t *Twitter) getAlbum(media []twitterMedia, fullText string) tb.Album {
	var file tb.Sendable
	var album = tb.Album{}
	var caption string

	for i, m := range media {
		if i == len(media)-1 {
			caption = fullText
		}

		if m.Type == "video" {
			file = &tb.Video{File: tb.FromURL(m.VideoInfo.best().URL), Caption: caption}
		} else if m.Type == "photo" {
			file = &tb.Photo{File: tb.FromURL(m.MediaURL), Caption: caption}
		} else {
			logger.Errorf("Unknown type: %s", m.Type)
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
	re := regexp.MustCompile(`\s?http\S+$`)
	text := re.ReplaceAllString(r.FullText, "")
	return fmt.Sprintf(`<a href="https://twitter.com/%s/status/%s">🐦</a> <b>%s</b> <i>(by %s)</i>\n%s`, r.User.ScreenName, r.ID, r.User.Name, m.Sender.Username, text)
}
