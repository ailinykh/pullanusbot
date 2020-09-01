package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// PlainLink to video url's processing
type PlainLink struct {
}

func (l *PlainLink) handleTextMessage(m *tb.Message) {
	b, ok := bot.(*tb.Bot)
	if !ok {
		logger.Error("Bot cast failed")
		return
	}

	r, _ := regexp.Compile(`^http(\S+)$`)
	if r.MatchString(m.Text) {
		logger.Infof("link found %s", m.Text)
		resp, err := http.Get(m.Text)
		if err != nil {
			logger.Errorf("%v", err)
		}
		switch resp.Header["Content-Type"][0] {
		case "video/mp4":
			b.Notify(m.Chat, tb.UploadingVideo)
			logger.Infof("found mp4 file %s", m.Text)
			video := &tb.Video{File: tb.FromURL(resp.Request.URL.String())}
			video.Caption = fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, m.Text, path.Base(resp.Request.URL.Path), m.Sender.Username)
			_, err := video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})

			if err == nil {
				logger.Info("Message sent. Deleting original")
				err = b.Delete(m)
				if err != nil {
					logger.Errorf("Can't delete original message: %v", err)
				}
			} else {
				logger.Errorf("Can't send entry: %v", err)

				if strings.HasPrefix(fmt.Sprint(err), "api error: Bad Request:") {
					logger.Info("Looks like Telegram API error. Trying upload as file...")
					l.sendByUploading(b, m)
				} else {
					b.Send(m.Chat, fmt.Sprint(err), &tb.SendOptions{ReplyTo: m})
				}
			}
		case "video/webm":
			l.sendByUploading(b, m)
		default:
			logger.Errorf("Unknown content type: %s", resp.Header["Content-Type"])
		}
	}
}

func (l *PlainLink) sendByUploading(b *tb.Bot, m *tb.Message) {
	resp, err := http.Get(m.Text)
	if err != nil {
		logger.Errorf("%v", err)
	}

	b.Notify(m.Chat, tb.UploadingVideo)
	filename := path.Base(resp.Request.URL.Path)
	videoFileSrc := path.Join(os.TempDir(), filename)
	videoFileDest := path.Join(os.TempDir(), filename+".mp4")
	defer os.Remove(videoFileSrc)
	defer os.Remove(videoFileDest)

	// logger.Printf("file %s, thumb: %s", videoFileSrc, videoThumbFile)

	// Download webm
	logger.Infof("downloading file %s", filename)
	err = downloadFile(videoFileSrc, m.Text)
	if err != nil {
		logger.Errorf("video download error: %v", err)
		return
	}

	// Convert webm to mp4
	logger.Infof("converting file %s", filename)
	cmd := fmt.Sprintf(`ffmpeg -y -i "%s" "%s"`, videoFileSrc, videoFileDest)
	_, err = exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Errorf("Video converting error: %v", err)
		return
	}
	logger.Infof("file converted!")

	c := Converter{}
	ffpInfo, err := c.getFFProbeInfo(videoFileDest)
	if err != nil {
		logger.Errorf("FFProbe info retreiving error: %v", err)
		return
	}

	videoStreamInfo, err := ffpInfo.getVideoStream()
	if err != nil {
		logger.Errorf("%v", err)
		return
	}

	video := tb.Video{File: tb.FromDisk(videoFileDest)}
	video.Width = videoStreamInfo.Width
	video.Height = videoStreamInfo.Height
	video.Duration = ffpInfo.Format.duration()
	video.SupportsStreaming = true
	video.Caption = fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, m.Text, filename, m.Sender.Username)

	// Getting thumbnail
	thumb, err := c.getThumbnail(videoFileDest)
	if err != nil {
		logger.Errorf("Thumbnail error: %v", err)
	} else {
		video.Thumbnail = &tb.Photo{File: tb.FromDisk(thumb)}
		defer os.Remove(thumb)
	}

	logger.Infof("Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

	_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err == nil {
		logger.Info("Video sent. Deleting original")
		err = b.Delete(m)
		if err != nil {
			logger.Errorf("Can't delete original message: %v", err)
		}
	} else {
		logger.Errorf("Can't send video: %v", err)
		b.Send(m.Chat, fmt.Sprint(err), &tb.SendOptions{ReplyTo: m})
	}
}
