package twitter

import (
	"fmt"
	"math"
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
	bot    i.Bot
	helper IHelper
)

// Twitter ...
type Twitter struct {
}

// Setup all nesessary command handlers
func (*Twitter) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	helper = Helper{}
	logger.Info("successfully initialized")
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

	tweet, err := helper.getTweet(tweetID)
	if err != nil {
		return
	}

	if len(tweet.Errors) > 0 {
		tweet.ID = tweetID // In case of timeout
		t.processError(tweet, m)
		return
	}

	if len(tweet.ExtendedEntities.Media) == 0 && tweet.QuotedStatus != nil && len(tweet.QuotedStatus.ExtendedEntities.Media) > 0 {
		tweet = *tweet.QuotedStatus
		logger.Warningf("tweet media is empty, using QuotedStatus instead %s", tweet.ID)
	}

	media := tweet.ExtendedEntities.Media

	switch len(media) {
	case 0:
		t.sendText(tweet, m)
	case 1:
		if media[0].Type == "video" || media[0].Type == "animated_gif" {
			t.sendVideo(media[0], tweet, m)
		} else if media[0].Type == "photo" {
			t.sendPhoto(media[0], tweet, m)
		} else {
			// Should not be there
			logger.Errorf("Unknown type: %s", media[0].Type)
			bot.Send(m.Chat, fmt.Sprintf("Unknown type: %s", media[0].Type), &tb.SendOptions{ReplyTo: m})
		}
	default:
		t.sendAlbum(media, tweet, m)
	}
}

func (t *Twitter) processError(tweet Tweet, m *tb.Message) {
	if tweet.Errors[0].Code == 88 { // "Rate limit exceeded 88"
		limit, err := strconv.ParseInt(tweet.Header["X-Rate-Limit-Reset"][0], 10, 64)
		if err != nil {
			logger.Error(err)
			return
		}
		timeout := limit - time.Now().Unix()
		logger.Infof("Twitter api timeout %d seconds", timeout)
		timeout = int64(math.Max(float64(timeout), 1)) // Twitter api timeout might be negative
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			t.processTweet(tweet.ID, m)
		}()
		return
	}

	logger.Error(tweet.Errors)
	logger.Info(tweet.Header)
	bot.Send(m.Chat, fmt.Sprint(tweet.Errors), &tb.SendOptions{ReplyTo: m})
}

func (t *Twitter) sendText(tweet Tweet, m *tb.Message) {
	logger.Info("Sending as text")
	caption := helper.makeCaption(m, tweet)
	for _, url := range tweet.Entities.Urls {
		caption += "\n" + url.ExpandedURL
	}
	_, err := bot.Send(m.Chat, caption, &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true})
	if err != nil {
		logger.Error(err)
	} else {
		t.deleteMessage(m)
	}
}

func (t *Twitter) sendPhoto(media Media, tweet Tweet, m *tb.Message) {
	file := &tb.Photo{File: tb.FromURL(media.MediaURL)}
	file.Caption = helper.makeCaption(m, tweet)
	logger.Infof("Sending as Photo %s", file.FileURL)
	bot.Notify(m.Chat, tb.UploadingPhoto)
	_, err := bot.Send(m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		logger.Error(err)
	} else {
		t.deleteMessage(m)
	}
}

func (t *Twitter) sendAlbum(media []Media, tweet Tweet, m *tb.Message) {
	logger.Infof("Sending as Album")
	caption := helper.makeCaption(m, tweet)
	bot.Notify(m.Chat, tb.UploadingPhoto)
	_, err := bot.SendAlbum(m.Chat, helper.makeAlbum(media, caption))
	if err != nil {
		logger.Error(err)
	} else {
		t.deleteMessage(m)
	}
}

func (t *Twitter) sendVideo(media Media, tweet Tweet, m *tb.Message) {
	file := &tb.Video{File: tb.FromURL(media.VideoInfo.best().URL)}
	caption := helper.makeCaption(m, tweet)
	file.Caption = caption
	logger.Infof("Sending as Video %s", file.FileURL)
	bot.Notify(m.Chat, tb.UploadingVideo)
	_, err := bot.Send(m.Chat, file, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		if strings.Contains(err.Error(), "failed to get HTTP URL content") || strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") {
			// Try to upload file to telegram
			logger.Info("Sending by uploading")

			filename := path.Base(media.VideoInfo.best().URL)
			filepath := path.Join(os.TempDir(), filename)
			defer os.Remove(filepath)

			err = u.DownloadFile(filepath, media.VideoInfo.best().URL)
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
			videofile.Upload(bot, m, caption, c.UploadFinishedCallback)
		} else {
			logger.Error(err)
		}
	} else {
		t.deleteMessage(m)
	}
}

func (Twitter) deleteMessage(m *tb.Message) {
	err := bot.Delete(m)
	if err != nil {
		logger.Warningf("Can't delete original message: %v", err)
	}
}
