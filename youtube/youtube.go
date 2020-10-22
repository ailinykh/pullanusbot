package youtube

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	c "pullanusbot/converter"
	i "pullanusbot/interfaces"
	"regexp"
	"strings"
	"sync"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Youtube to video url's processing
type Youtube struct {
	mutex sync.Mutex
}

// Video is a struct to handle youtube-dl's JSON output
type Video struct {
	ID       string   `json:"id"`
	Duration int      `json:"duration"`
	Formats  []Format `json:"formats"`
	Title    string   `json:"title"`
}

// Format is a description of available formats for downloading
type Format struct {
	Ext      string `json:"ext"`
	Filesize int    `json:"filesize"`
	FormatID string `json:"format_id"`
}

// Setup all nesessary command handlers
func (y *Youtube) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle("/yt", y.processMessage)
	logger.Info("Successfully initialized")
}

// HandleTextMessage is an i.TextMessageHandler interface implementation
func (y *Youtube) HandleTextMessage(m *tb.Message) {
	r := regexp.MustCompile(`https?:\S+youtu[\.be|\.com]\S+`)
	match := r.FindStringSubmatch(m.Text)
	if len(match) > 0 {
		y.processURL(match[0], m)
	}
}

func (y *Youtube) processMessage(m *tb.Message) {
	y.processURL(m.Payload, m)
}

func (y *Youtube) processURL(url string, m *tb.Message) {
	logger.Infof("Processing url: %s", url)

	cmd := fmt.Sprintf(`youtube-dl -j "%s"`, url)
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Error(err)
		logger.Info(out)
		return
	}

	var video Video
	err = json.Unmarshal(out, &video)
	if err != nil {
		logger.Error(err)
		return
	}

	const SizeLimit = 900 // 15 min

	if len(m.Payload) > 0 {
		logger.Infof("Uploading by payload: %s", m.Payload)
		y.uploadVideo(video, m)
	} else if video.Duration < SizeLimit {
		logger.Infof("Uploading by duration: %d", video.Duration)
		y.uploadVideo(video, m)
	} else {
		logger.Infof("File longer than %d sec (%d). Skipping...", SizeLimit, video.Duration)
	}
}

func (y *Youtube) uploadVideo(video Video, m *tb.Message) {
	// Just one video at one time pls
	y.mutex.Lock()
	defer y.mutex.Unlock()

	filepath := path.Join(os.TempDir(), "youtube-"+video.ID+".mp4")
	defer os.Remove(filepath)

	cmd := fmt.Sprintf("youtube-dl -f 134+140 %s -o %s", video.ID, filepath)
	logger.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	err := exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		logger.Error(err)
		return
	}

	videoFile, err := c.NewVideoFile(filepath)
	if err != nil {
		logger.Errorf("Can't create video file for %s, %v", filepath, err)
		return
	}
	defer videoFile.Dispose()

	const SizeLimit = 50000000
	if videoFile.Size > SizeLimit {
		logger.Infof("file is over 50MB, duration: %d", video.Duration)
		duration, n := 0, 0
		var videoFiles = []*c.VideoFile{}
		for duration < videoFile.Duration() {
			nextFilePath := fmt.Sprintf("%s-%d.mp4", filepath, n)
			cmd := fmt.Sprintf(`ffmpeg -i %s -ss %d -fs %d %s`, filepath, duration, SizeLimit, nextFilePath)
			logger.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
			out, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				logger.Error(out, " - ", err)
				return
			}
			nextVideoFile, err := c.NewVideoFile(nextFilePath)
			if err != nil {
				logger.Errorf("Can't create next video file for %s, %v", nextFilePath, err)
				return
			}
			defer nextVideoFile.Dispose()

			videoFiles = append(videoFiles, nextVideoFile)
			duration += nextVideoFile.Duration()
			n++
		}

		for idx, vf := range videoFiles {
			caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>[%d/%d] %s</b> <i>(by %s)</i>`, video.ID, idx+1, len(videoFiles), video.Title, m.Sender.Username)
			vf.Upload(bot, m, caption, func(i.Bot, *tb.Message) {})
		}
		c.UploadFinishedCallback(bot, m)

	} else {
		caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, video.ID, video.Title, m.Sender.Username)
		videoFile.Upload(bot, m, caption, c.UploadFinishedCallback)
	}
}
