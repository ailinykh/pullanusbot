package link

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"pullanusbot/converter"
	c "pullanusbot/converter"
	i "pullanusbot/interfaces"
	u "pullanusbot/utils"
	"regexp"
	"sync"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Link searches video links inside text messages
type Link struct {
	mutex sync.Mutex
}

// Setup all nesessary command handlers
func (l *Link) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	logger.Info("Successfully initialized")
}

// HandleTextMessage is an i.TextMessageHandler interface implementation
func (l *Link) HandleTextMessage(m *tb.Message) {
	r, _ := regexp.Compile(`^http(\S+)$`)
	if r.MatchString(m.Text) {
		l.processLink(m.Text, m)
	}
	// r := regexp.MustCompile(`http\S+`)
	// for _, link := range r.FindAllString(m.Text, -1) {
	// 	l.processLink(link, m)
	// }
}

func (l *Link) processLink(link string, m *tb.Message) {
	logger.Infof("Link found: %s", link)
	resp, err := http.Get(link)

	if err != nil {
		logger.Error(err)
		return
	}

	switch resp.Header["Content-Type"][0] {
	case "video/mp4":
		bot.Notify(m.Chat, tb.UploadingVideo)
		logger.Infof("found mp4 file %s", link)
		video := &tb.Video{File: tb.FromURL(resp.Request.URL.String())}
		video.Caption = fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, link, path.Base(resp.Request.URL.Path), m.Sender.Username)
		_, err := video.Send(bot.(*tb.Bot), m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})

		if err == nil {
			logger.Info("Message sent. Deleting original")
			err = bot.Delete(m)
			if err != nil {
				logger.Error(err)
			}
		} else {
			logger.Error(err)
			// telegram error, fallback to upload
			l.downloadAndSend(link, m)
		}
	case "video/webm":
		l.downloadAndSend(link, m)
	default:
		logger.Warningf("Unsupported content type: %s", resp.Header["Content-Type"])
	}
}

func (l *Link) downloadAndSend(link string, m *tb.Message) {
	// Just one video at one time pls
	l.mutex.Lock()
	defer l.mutex.Unlock()

	resp, err := http.Get(link)
	if err != nil {
		logger.Error(err)
	}

	bot.Notify(m.Chat, tb.UploadingVideo)
	filename := path.Base(resp.Request.URL.Path)
	srcPath := path.Join(os.TempDir(), filename)
	dstPath := path.Join(os.TempDir(), filename+".mp4")
	defer os.Remove(srcPath)
	defer os.Remove(dstPath)

	// Download webm
	logger.Infof("downloading file %s", filename)
	err = u.DownloadFile(srcPath, link)
	if err != nil {
		logger.Error(err)
		return
	}

	// Convert webm to mp4
	logger.Infof("converting file %s", srcPath)
	cmd := fmt.Sprintf("ffmpeg -v error -y -i %s %s", srcPath, dstPath)
	_, err = exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logger.Error(err)
		return
	}

	videofile, err := c.NewVideoFile(dstPath)
	if err != nil {
		logger.Errorf("Can't create video file for %s, %v", dstPath, err)
		return
	}
	defer videofile.Dispose()
	caption := fmt.Sprintf(`<a href="%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, link, filename, m.Sender.Username)
	videofile.Upload(bot, m, caption, converter.UploadFinishedCallback)
}
