package api

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/use_cases"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (t *Telebot) SetupVideo(video_flow *use_cases.VideoFlow) {
	var mutex sync.Mutex

	t.bot.Handle(tb.OnDocument, func(m *tb.Message) {
		if m.Document.MIME[:5] == "video" || m.Document.MIME == "image/gif" {
			mutex.Lock()
			defer mutex.Unlock()

			t.logger.Infof("Attempt to download %s %s (sent by %s)", m.Document.FileName, m.Document.MIME, m.Sender.Username)

			path := path.Join(os.TempDir(), m.Document.FileName)
			err := t.bot.Download(&m.Document.File, path)
			if err != nil {
				t.logger.Error(err)
				return
			}

			t.logger.Infof("Downloaded to %s", path)
			defer os.Remove(path)
			inputFile, err := t.videoFileFactory.CreateVideoFile(path)

			if err != nil {
				t.logger.Error(err)
				return
			}

			defer os.Remove(inputFile.ThumbPath)

			outputFile, err := video_flow.Process(inputFile)

			if err != nil {
				chatID, e := strconv.ParseInt(os.Getenv("ADMIN_CHAT_ID"), 10, 64)
				if e == nil {
					chat := &tb.Chat{ID: chatID}
					t.bot.Forward(chat, m)
					t.bot.Send(chat, err.Error())
				}
				return
			}

			defer os.Remove(outputFile.FilePath)
			defer os.Remove(outputFile.ThumbPath)

			caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", m.Document.FileName, m.Sender.Username)
			if inputFile.FilePath != outputFile.FilePath {
				fi, _ := os.Stat(outputFile.FilePath)
				caption = fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, inputFile.Bitrate/1024, float32(fi.Size())/1048576, outputFile.Bitrate/1024)
			}
			video := makeVideoFile(outputFile, caption)
			t.bot.Notify(m.Chat, tb.UploadingVideo)
			_, err = video.Send(t.bot, m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
			if err != nil {
				t.logger.Error(err)
			} else {
				t.logger.Infof("%s sent successfully", outputFile.FileName)
				t.bot.Delete(m)
			}
		} else {
			t.logger.Infof("%s not supported yet", m.Document.MIME)
		}
	})

	t.bot.Handle(tb.OnAnimation, func(m *tb.Message) {
		t.logger.Info(m.Animation)
	})
}

func makeVideoFile(vf *core.VideoFile, caption string) tb.Video {
	video := tb.Video{File: tb.FromDisk(vf.FilePath)}
	video.Width = vf.Width
	video.Height = vf.Height
	video.Caption = caption
	video.Duration = vf.Duration
	video.SupportsStreaming = true
	video.Thumbnail = &tb.Photo{File: tb.FromDisk(vf.ThumbPath)}
	return video
}
