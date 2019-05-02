package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"

	tb "gopkg.in/tucnak/telebot.v2"
)

// PlainLink to video url's processing
type PlainLink struct {
}

func (l *PlainLink) handleTextMessage(m *tb.Message) {
	b, ok := bot.(*tb.Bot)
	if !ok {
		log.Println("PlainLink: Bot cast failed")
		return
	}

	r, _ := regexp.Compile(`^http(\S+)$`)
	if r.MatchString(m.Text) {
		log.Printf("PlainLink: link found %s", m.Text)
		resp, err := http.Get(m.Text)
		if err != nil {
			log.Printf("PlainLink: %s", err)
		}
		switch resp.Header["Content-Type"][0] {
		case "video/mp4":
			log.Printf("PlainLink: found mp4 file %s", m.Text)
			video := &tb.Video{File: tb.FromURL(resp.Request.URL.String())}
			video.Caption = fmt.Sprintf("[🎞](%s) *%s* _(by %s)_", m.Text, path.Base(resp.Request.URL.Path), m.Sender.Username)
			_, err := video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})

			if err == nil {
				log.Println("PlainLink: Message sent. Deleting original")
				err = b.Delete(m)
				if err != nil {
					log.Printf("PlainLink: Can't delete original message: %s", err)
				}
			} else {
				log.Printf("PlainLink: Can't send entry: %s", err)
			}
		case "video/webm":
			filename := path.Base(resp.Request.URL.Path)
			videoFileSrc := path.Join(os.TempDir(), filename)
			videoFileDest := path.Join(os.TempDir(), filename+".mp4")
			defer os.Remove(videoFileSrc)
			defer os.Remove(videoFileDest)

			// log.Printf("PlainLink: file %s, thumb: %s", videoFileSrc, videoThumbFile)

			// Download webm
			log.Printf("PlainLink: downloading file %s", filename)
			err = downloadFile(videoFileSrc, m.Text)
			if err != nil {
				log.Printf("PlainLink: video download error: %s", err)
				return
			}

			// Convert webm to mp4
			log.Printf("PlainLink: converting file %s", filename)
			cmd := fmt.Sprintf(`ffmpeg -y -i "%s" "%s"`, videoFileSrc, videoFileDest)
			_, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				log.Printf("PlainLink: Video converting error: %s", err)
				return
			}
			log.Println("PlainLink: file converted!")

			c := Converter{}
			ffpInfo, err := c.getFFProbeInfo(videoFileDest)
			if err != nil {
				log.Printf("PlainLink: FFProbe info retreiving error: %s", err)
				return
			}

			videoStreamInfo, err := ffpInfo.getVideoStream()
			if err != nil {
				log.Printf("PlainLink: %s", err)
				return
			}

			video := tb.Video{File: tb.FromDisk(videoFileDest)}
			video.Width = videoStreamInfo.Width
			video.Height = videoStreamInfo.Height
			video.Duration = ffpInfo.Format.duration()
			video.SupportsStreaming = true
			video.Caption = fmt.Sprintf("[🎞](%s) *%s* _(by %s)_", m.Text, filename, m.Sender.Username)

			// Getting thumbnail
			thumb, err := c.getThumbnail(videoFileDest)
			if err != nil {
				log.Printf("PlainLink: Thumbnail error: %s", err)
			} else {
				video.Thumbnail = &tb.Photo{File: tb.FromDisk(thumb)}
				defer os.Remove(thumb)
			}

			log.Printf("PlainLink: Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

			_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			if err == nil {
				log.Println("PlainLink: Video sent. Deleting original")
				err = b.Delete(m)
				if err != nil {
					log.Printf("PlainLink: Can't delete original message: %s", err)
				}
			} else {
				log.Printf("PlainLink: Can't send video: %s", err)
			}
		default:
			log.Printf("PlainLink: Unknown content type: %s", resp.Header["Content-Type"])
		}
	}
}
