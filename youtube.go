package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Youtube to video url's processing
type Youtube struct {
	mutex sync.Mutex
}

// YoutubeVideo is a struct to handle youtube-dl's JSON output
type YoutubeVideo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (y *Youtube) initialize() {
	bot.Handle("/yt", y.processMessage)
	logger.Info("successfully initialized")
}

func (y *Youtube) processMessage(m *tb.Message) {
	// Just one video at one time pls
	y.mutex.Lock()
	defer y.mutex.Unlock()

	logger.Info("Got payload: ", m.Payload)

	filepath := path.Join(os.TempDir(), randStringRunes(12)+".mp4")
	cmd := fmt.Sprintf(`youtube-dl --print-json -f 134+140 "%s" -o "%s"`, m.Payload, filepath)
	logger.Info(cmd)
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Error(out, " - ", err)
		return
	}

	var ytVideo YoutubeVideo
	err = json.Unmarshal(out, &ytVideo)
	if err != nil {
		logger.Error(err)
		return
	}

	// defer os.Remove(filename)

	videoFile, err := NewVideoFile(filepath)
	if err != nil {
		logger.Errorf("Can't create video file for %s, %v", filepath, err)
		return
	}
	defer os.Remove(videoFile.filepath)
	defer os.Remove(videoFile.thumbpath)

	const SizeLimit = 50000000
	if videoFile.size > SizeLimit {
		logger.Info("file is over 50MB")
		duration, n := 0, 0
		var videoFiles = []*VideoFile{}
		for duration < videoFile.ffpInfo.Format.duration() {
			nextFilePath := fmt.Sprintf("%s_%d.mp4", filepath, n)
			cmd := fmt.Sprintf(`ffmpeg -i %s -ss %d -fs %d %s`, filepath, duration, SizeLimit, nextFilePath)
			logger.Info(cmd)
			out, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				logger.Error(out, " - ", err)
				return
			}
			nextVideoFile, err := NewVideoFile(nextFilePath)
			if err != nil {
				logger.Errorf("Can't create next video file for %s, %v", nextFilePath, err)
				return
			}
			defer os.Remove(nextVideoFile.filepath)
			defer os.Remove(nextVideoFile.thumbpath)

			videoFiles = append(videoFiles, nextVideoFile)
			duration += nextVideoFile.ffpInfo.Format.duration()
			n++
		}

		for i, vf := range videoFiles {
			caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>[%d/%d] %s</b> <i>(by %s)</i>`, ytVideo.ID, i+1, len(videoFiles), ytVideo.Title, m.Sender.Username)
			uploadFile(vf, m, caption)
		}

	} else {
		caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, ytVideo.ID, ytVideo.Title, m.Sender.Username)
		uploadFile(videoFile, m, caption)
	}
}
