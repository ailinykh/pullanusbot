package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	c "pullanusbot/converter"
	i "pullanusbot/interfaces"
	u "pullanusbot/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Twitter ...
type Twitter struct {
}

// Setup all nesessary command handlers
func (*Twitter) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	logger.Info("Successfully initialized")
}

// HandleTextMessage is an i.TextMessageHandler interface implementation
func (t *Twitter) HandleTextMessage(m *tb.Message) {
	r, _ := regexp.Compile(`twitter\.com.+/(\d+)\S*$`)
	match := r.FindStringSubmatch(m.Text)
	if len(match) > 1 {
		t.processTweet(match[1], m)
	}
}

func (t *Twitter) processTweet(tweetID string, m *tb.Message) {
	logger.Infof("Processing tweet %s", tweetID)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", tweetID), nil)
	req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAPYXBAAAAAAACLXUNDekMxqa8h%2F40K4moUkGsoc%3DTYfbDKbT3jJPCEVnMYqilB28NHfOPqkca3qaAxGfsyKCs0wRbw")
	res, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer res.Body.Close()

	var twResp twitterReponse
	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(body, &twResp)
	if err != nil {
		logger.Error(err)
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
			timeout = int64(math.Max(float64(timeout), 3)) // Twitter api timeout might be negative

			go func() {
				time.Sleep(time.Duration(timeout) * time.Second)
				t.processTweet(tweetID, m)
			}()

			return
		}

		logger.Error(twResp.Errors)
		logger.Info(res.Header)

		bot.Send(m.Chat, fmt.Sprint(twResp.Errors), &tb.SendOptions{ReplyTo: m})
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

		_, err = bot.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true})
	case 1:
		if media[0].Type == "video" {
			file := &tb.Video{File: tb.FromURL(media[0].VideoInfo.best().URL)}
			file.Caption = caption
			logger.Infof("Sending as Video %s", file.FileURL)
			bot.Notify(m.Chat, tb.UploadingVideo)
			_, err = file.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
		} else if media[0].Type == "photo" {
			file := &tb.Photo{File: tb.FromURL(media[0].MediaURL)}
			file.Caption = caption
			logger.Infof("Sending as Photo %s", file.FileURL)
			bot.Notify(m.Chat, tb.UploadingPhoto)
			_, err = file.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
		} else {
			logger.Infof("Unknown type: %s", media[0].Type)
			bot.Send(m.Chat, fmt.Sprintf("Unknown type: %s", media[0].Type), &tb.SendOptions{ReplyTo: m})
			return
		}
	default:
		logger.Infof("Sending as Album")
		bot.Notify(m.Chat, tb.UploadingPhoto)
		_, err = bot.SendAlbum(m.Chat, t.getAlbum(media, caption, twResp.FullText))
	}

	if err == nil {
		logger.Info("Messages sent. Deleting original")
		err = bot.Delete(m)
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

			err = u.DownloadFile(filepath, media[0].VideoInfo.best().URL)
			if err != nil {
				logger.Errorf("video download error: %v", err)
				return
			}

			videofile, err := c.NewVideoFile(filepath)
			if err != nil {
				logger.Errorf("Can't create video file for %s, %v", filepath, err)
				return
			}
			defer videofile.Dispose()
			videofile.Upload(bot, m, caption)
		}
	}
}

func (t *Twitter) getAlbum(media []twitterMedia, photoCaption, videoCaption string) tb.Album {
	var file tb.Sendable
	var album = tb.Album{}
	var pc, vc string

	for i, m := range media {
		if i == len(media)-1 {
			vc = videoCaption
			pc = photoCaption
		}

		if m.Type == "video" {
			file = &tb.Video{File: tb.FromURL(m.VideoInfo.best().URL), Caption: vc}
		} else if m.Type == "photo" {
			file = &tb.Photo{File: tb.FromURL(m.MediaURL), Caption: pc, ParseMode: tb.ModeHTML}
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
	return fmt.Sprintf("<a href='https://twitter.com/%s/status/%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", r.User.ScreenName, r.ID, r.User.Name, m.Sender.Username, text)
}
