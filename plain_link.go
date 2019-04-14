package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	tb "gopkg.in/tucnak/telebot.v2"
)

// PlainLink to video url's processing
type PlainLink struct {
}

func (l *PlainLink) handleTextMessage(m *tb.Message) {
	// b, ok := bot.(*tb.Bot)
	// if !ok {
	// 	log.Println("PlainLink: Bot cast failed")
	// 	return
	// }

	r, _ := regexp.Compile(`^http(\S+)$`)
	if r.MatchString(m.Text) {
		resp, err := http.Get(m.Text)
		if err != nil {
			log.Printf("PlainLink: %s", err)
		}
		switch resp.Header["Content-Type"][0] {
		case "video/mp4":
			log.Printf("PlainLink: found mp4 file %s", m.Text)
			l.sendMP4Video(m, m.Text)
		// case "video/webm":
		// 	link := strings.Replace(m.Text, ".webm", ".mp4", 1)
		// 	log.Printf("PlainLink: checking webm %s", link)
		// 	resp, err = http.Get(link)
		// 	if err != nil {
		// 		log.Printf("PlainLink: webm %s", err)
		// 	}

		// 	if resp.Status == "200 OK" {
		// 		// TODO: file name
		// 		videoFile := path.Join(os.TempDir(), "plainlink_video.mp4")
		// 		videoThumbFile := path.Join(os.TempDir(), "plainlink_video_thumb.jpg")
		// 		defer os.Remove(videoFile)
		// 		defer os.Remove(videoThumbFile)

		// 		// log.Printf("PlainLink: file %s, thumb: %s", videoFile, videoThumbFile)

		// 		err = l.downloadFile(videoFile, link)
		// 		if err != nil {
		// 			log.Printf("PlainLink: video download error: %s", err)
		// 		}

		// 		c := Converter{}
		// 		ffpInfo, err := c.getFFProbeInfo(videoFile)
		// 		if err != nil {
		// 			log.Printf("PlainLink: FFProbe info retreiving error: %s", err)
		// 			return
		// 		}

		// 		videoStreamInfo, err := ffpInfo.getVideoStream()
		// 		if err != nil {
		// 			log.Printf("PlainLink: %s", err)
		// 			return
		// 		}

		// 		video := tb.Video{File: tb.FromDisk(videoFile)}

		// 		scale := "90:-1"
		// 		if videoStreamInfo.Width < videoStreamInfo.Height {
		// 			scale = "-1:90"
		// 		}

		// 		video.Width = videoStreamInfo.Width
		// 		video.Height = videoStreamInfo.Height
		// 		video.Duration = ffpInfo.Format.duration()
		// 		video.SupportsStreaming = true
		// 		video.Caption = fmt.Sprintf("[🎞](%s) _(by %s)_", link, m.Sender.Username)

		// 		// Getting thumbnail
		// 		cmd := fmt.Sprintf("ffmpeg -i \"%s\" -ss 00:00:01.000 -vframes 1 -filter:v scale=\"%s\" \"%s\"", videoFile, scale, videoThumbFile)
		// 		err = exec.Command("/bin/sh", "-c", cmd).Run()
		// 		if err != nil {
		// 			log.Printf("PlainLink: Thumbnail error: %s", err)
		// 		} else {
		// 			thumb := tb.Photo{File: tb.FromDisk(videoThumbFile)}
		// 			video.Thumbnail = &thumb
		// 		}

		// 		log.Printf("PlainLink: Sending file: w:%d h:%d duration:%d", video.Width, video.Height, video.Duration)

		// 		_, err = video.Send(b, m.Chat, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		// 		if err == nil {
		// 			log.Println("PlainLink: Video sent. Deleting original")
		// 			err = b.Delete(m)
		// 			if err != nil {
		// 				log.Printf("PlainLink: Can't delete original message: %s", err)
		// 			}
		// 		} else {
		// 			log.Printf("PlainLink: Can't send video: %s", err)
		// 		}

		// 	} else {
		// 		log.Println("PlainLink: Wrong webm response")
		// 		for name, values := range resp.Header {
		// 			for _, value := range values {
		// 				log.Println("\t", name, value)
		// 			}
		// 		}
		// 	}
		default:
			log.Printf("PlainLink: Unknown content type: %s", resp.Header["Content-Type"])
		}
	}
}

func (l *PlainLink) sendMP4Video(m *tb.Message, link string) {
	b, ok := bot.(*tb.Bot)
	if !ok {
		log.Println("PlainLink: Bot cast failed")
		return
	}

	video := &tb.Video{File: tb.FromURL(link)}
	video.Caption = fmt.Sprintf("[🎞](%s) _(by %s)_", link, m.Sender.Username)
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
}

func (l *PlainLink) downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
