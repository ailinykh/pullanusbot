package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Twitter goes to prettify twitter links
type Twitter struct {
}

type twitterReponse struct {
	ID               string          `json:"id_str"`
	FullText         string          `json:"full_text"`
	ExtendedEntities twitterEntity   `json:"extended_entities,omitempty"`
	User             twitterUser     `json:"user"`
	QuotedStatus     *twitterReponse `json:"quoted_status,omitempty"`
}

type twitterUser struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type twitterEntity struct {
	Media []twitterMedia `json:"media"`
}

type twitterMedia struct {
	MediaURL      string           `json:"media_url"`
	MediaURLHTTPS string           `json:"media_url_https"`
	Type          string           `json:"type"`
	VideoInfo     twitterVideoInfo `json:"video_info"`
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

func (t *Twitter) handleTextMessage(m *tb.Message) {
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
			log.Printf("Twitter: json fetch error: %s", err)
			return
		}
		defer res.Body.Close()

		var twResp twitterReponse
		body, _ := ioutil.ReadAll(res.Body)

		err = json.Unmarshal(body, &twResp)
		if err != nil {
			log.Printf("Twitter: json parse error: %s", err)
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
			log.Println("Twitter: Senting as text")
			_, err = b.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeMarkdown, DisableWebPagePreview: true})
			log.Printf("Twitter: %v", err)
		case 1:
			if media[0].Type == "video" {
				file := &tb.Video{File: tb.FromURL(media[0].VideoInfo.best().URL)}
				file.Caption = caption
				log.Printf("Twitter: Sending as Video %s", file.FileURL)
				_, err = file.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			} else if media[0].Type == "photo" {
				file := &tb.Photo{File: tb.FromURL(media[0].MediaURL)}
				file.Caption = caption
				log.Printf("Twitter: Sending as Photo %s", file.FileURL)
				_, err = file.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			} else {
				log.Printf("Twitter: Unknown type: %s", media[0].Type)
			}
		default:
			log.Printf("Twitter: Sending as Album")
			b.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeMarkdown, DisableWebPagePreview: true})
			_, err = b.SendAlbum(m.Chat, t.getAlbum(media))
		}

		if err == nil {
			log.Println("Twitter: Messages sent. Deleting original")
			err = b.Delete(m)
			if err != nil {
				log.Printf("Twitter: Can't delete original message: %s", err)
			}
		} else {
			log.Printf("Twitter: Can't send entry: %s", err)

			if strings.HasSuffix(err.Error(), "failed to get HTTP URL content") {
				// Try to upload file to telegram
				log.Println("Twitter: Sending by uploading")

				filename := path.Base(media[0].VideoInfo.best().URL)
				videoFile := path.Join(os.TempDir(), filename)
				defer os.Remove(videoFile)

				err = downloadFile(videoFile, media[0].VideoInfo.best().URL)
				if err != nil {
					log.Printf("Twitter: video download error: %s", err)
					return
				}

				c := Converter{}
				ffpInfo, err := c.getFFProbeInfo(videoFile)
				if err != nil {
					log.Printf("Twitter: FFProbe info retreiving error: %s", err)
					return
				}

				videoStreamInfo, err := ffpInfo.getVideoStream()
				if err != nil {
					log.Printf("Twitter: %s", err)
					return
				}

				video := tb.Video{File: tb.FromDisk(videoFile)}
				video.Width = videoStreamInfo.Width
				video.Height = videoStreamInfo.Height
				video.Duration = ffpInfo.Format.duration()
				video.SupportsStreaming = true
				// insert hot link
				idx := strings.Index(caption, " ")

				video.Caption = caption[0:idx] + fmt.Sprintf("[🎞](%s)", media[0].VideoInfo.best().URL) + caption[idx:]

				// Getting thumbnail
				thumb, err := c.getThumbnail(videoFile)
				if err != nil {
					log.Printf("PlainLink: Thumbnail error: %s", err)
				} else {
					video.Thumbnail = &tb.Photo{File: tb.FromDisk(thumb)}
					defer os.Remove(thumb)
				}

				log.Printf("Twitter: Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

				_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
				if err == nil {
					log.Println("Twitter: Video sent. Deleting original")
					err = b.Delete(m)
					if err != nil {
						log.Printf("Twitter: Can't delete original message: %s", err)
					}
				} else {
					log.Printf("Twitter: Can't send video: %s", err)
				}
			}
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
	re := regexp.MustCompile(`\s?http\S+$`)
	text := re.ReplaceAllString(r.FullText, "")
	return fmt.Sprintf("[🐦](https://twitter.com/%s/status/%s) *%s* _(by %s)_\n%s", r.User.ScreenName, r.ID, r.User.Name, m.Sender.Username, text)
}
