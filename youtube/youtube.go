package youtube

import (
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

// Setup all nesessary command handlers
func (y *Youtube) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle("/yt", y.processMessage)
	logger.Info("Successfully initialized")
}

// HandleTextMessage is an i.TextMessageHandler interface implementation
func (y *Youtube) HandleTextMessage(m *tb.Message) {
	r := regexp.MustCompile(`https?:\/\/youtu[\.be|\.com]\S+`)
	match := r.FindStringSubmatch(m.Text)
	if len(match) > 0 {
		if m.Private() {
			y.processURL(match[0], m)
		} else {
			video, err := getVideo(match[0])
			if err != nil {
				logger.Error(err)
				return
			}
			if video.Duration < 900 {
				y.uploadVideo(video, "134", m)
			}
		}
	}
}

func (y *Youtube) processMessage(m *tb.Message) {
	y.processURL(m.Payload, m)
}

func (y *Youtube) processURL(url string, m *tb.Message) {
	logger.Infof("Processing url: %s", url)
	video, err := getVideo(url)
	if err != nil {
		logger.Error(err)
		return
	}

	keyboard := [][]tb.InlineButton{}
	for _, f := range video.availableFormats() {
		mbSize := float64(f.Filesize+video.audioFormat().Filesize) / 1024 / 1024
		if mbSize > 2000 {
			logger.Warningf("size limit exceeded for %s %.02fMiB", f.FormatID, mbSize)
			continue
		}
		text := fmt.Sprintf("%dx%d (%s) - %.02fMiB", f.Width, f.Height, f.FormatNote, mbSize)
		logger.Info(text)
		btn := tb.InlineButton{Text: text, Unique: "_" + f.FormatID, Data: video.ID + "|" + f.FormatID}
		bot.Handle(&btn, y.handleDlCb)
		keyboard = append(keyboard, []tb.InlineButton{btn})
	}

	btn := tb.InlineButton{Text: "❌ Cancel", Unique: "cancel"}
	bot.Handle(&btn, y.handleDlCb)
	keyboard = append(keyboard, []tb.InlineButton{btn})

	caption := fmt.Sprintf("<b>%s</b>", video.Title)
	menu := &tb.ReplyMarkup{InlineKeyboard: keyboard}
	opts := &tb.SendOptions{ParseMode: tb.ModeHTML, ReplyMarkup: menu}
	thumb := video.thumb()
	file := &tb.Photo{File: tb.FromURL(thumb.URL), Caption: caption}
	_, err = file.Send(bot.(*tb.Bot), m.Chat, opts)
	// caption := fmt.Sprintf(`<a href="https://i.ytimg.com/vi/%s/maxresdefault.jpg"> </a><b>%s</b>`, video.ID, video.Title)
	// _, err = bot.Send(m.Chat, caption, opts)
	if err != nil {
		logger.Error(err)
		logger.Infof("upload image %s", thumb.URL)
		filepath := path.Join(os.TempDir(), "youtube-"+video.ID+".jpg")
		defer os.Remove(filepath)
		downloadFile(filepath, thumb.URL)
		file = &tb.Photo{File: tb.FromDisk(filepath), Width: thumb.Width, Height: thumb.Height, Caption: caption}
		_, err = file.Send(bot.(*tb.Bot), m.Chat, opts)
		if err != nil {
			logger.Error(err)
		}
	}
}

func (y *Youtube) handleDlCb(c *tb.Callback) {
	bot.Respond(c, &tb.CallbackResponse{})
	data := strings.Split(c.Data, "|")
	logger.Infof("%#v", data)

	if strings.Contains(data[0], "cancel") {
		logger.Info("cancel operation")
		err := bot.Delete(c.Message)
		if err != nil {
			logger.Error(err)
		}
		return
	}

	videoID, formatID := data[len(data)-2], data[len(data)-1]

	video, err := getVideo(videoID)
	if err != nil {
		logger.Error(err)
		bot.Edit(c.Message, &tb.Photo{File: c.Message.Photo.File, Caption: err.Error()})
		return
	}

	opts := &tb.SendOptions{ParseMode: tb.ModeHTML}
	f := video.formatByID(formatID)
	caption := fmt.Sprintf("<b>%s</b>\n\n<i>processing %s...</i>", video.Title, f.FormatNote)
	_, err = bot.Edit(c.Message, &tb.Photo{File: c.Message.Photo.File, Caption: caption}, opts)
	if err != nil {
		logger.Error(err)
		return
	}

	y.uploadVideo(video, formatID, c.Message)
}

func (y *Youtube) uploadVideo(video *Video, formatID string, m *tb.Message) {
	// Just one video at one time pls
	y.mutex.Lock()
	defer y.mutex.Unlock()

	logger.Infof("Uploading %s format %s", video.ID, formatID)

	filepath := path.Join(os.TempDir(), "youtube-"+video.ID+".mp4")
	defer os.Remove(filepath)

	cmd := fmt.Sprintf("youtube-dl -f %s+140 %s -o %s", formatID, video.ID, filepath)
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

	caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, video.ID, video.Title, m.Sender.Username)
	if m.Sender.IsBot {
		caption = fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> `, video.ID, video.Title)
	}

	err = videoFile.Upload(bot, m, caption, c.UploadFinishedCallback)
	if err != nil {
		if strings.Contains(err.Error(), "Request Entity Too Large") {
			const SizeLimit = 50000000
			logger.Infof("file size is %.02fMB, duration: %d", float64(videoFile.Size)/1024/1024, video.Duration)
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
				if m.Sender.IsBot {
					caption = fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>[%d/%d] %s</b>`, video.ID, idx+1, len(videoFiles), video.Title)
				}
				vf.Upload(bot, m, caption, func(i.Bot, *tb.Message) {})
			}
			c.UploadFinishedCallback(bot, m)
		}
	}
}
