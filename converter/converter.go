package converter

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"sync"

	i "pullanusbot/interfaces"

	"github.com/google/logger"
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var bot i.Bot

// Converter helps to post video files proper way
type Converter struct {
	mutex sync.Mutex
}

// Setup is a basic initialization method
func (c *Converter) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle(tb.OnDocument, c.checkMessage)
	logger.Info("Successfully initialized")
}

func (c *Converter) checkMessage(m *tb.Message) {
	if m.Document.MIME[:5] == "video" {
		// Just one video at one time pls
		c.mutex.Lock()
		defer c.mutex.Unlock()

		logger.Infof("Got video! \"%s\" of type %s from %s", m.Document.FileName, m.Document.MIME, m.Sender.Username)

		if m.Document.FileSize > 20*1024*1024 {
			logger.Errorf("File is greater than 20 MB :(%d)", m.Document.FileSize)
			return
		}

		srcPath := path.Join(os.TempDir(), m.Document.FileName)
		dstPath := path.Join(os.TempDir(), "converted_"+m.Document.FileName)
		defer os.Remove(srcPath)
		defer os.Remove(dstPath)

		logger.Info("Downloading video...")
		bot.Download(&m.Document.File, srcPath)
		logger.Info("Video downloaded")

		videofile, err := NewVideoFile(srcPath)
		if err != nil {
			logger.Errorf("Can't create video file for %s, %v", srcPath, err)
			return
		}
		defer videofile.Dispose()

		if videofile.ffpInfo.Format.NbStreams == 1 {
			logger.Error("Assuming gif file. Skipping...")
			return
		}

		srcBitrate := videofile.videoStreamInfo.bitrate()
		dstBitrate := int(math.Min(float64(srcBitrate), 568320))

		logger.Infof("Source file bitrate: %d, destination file bitrate: %d", srcBitrate, dstBitrate)

		if srcBitrate != dstBitrate {
			logger.Info("Bitrates not equal. Converting...")
			cmd := fmt.Sprintf(`ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, srcPath, dstBitrate/1024, srcPath, dstBitrate/1024, dstPath)
			err = exec.Command("/bin/sh", "-c", cmd).Run()
			if err != nil {
				logger.Errorf("Video converting error: %v", err)
				return
			}
			logger.Info("Video converted successfully")
		}

		fi, err := os.Stat(dstPath)
		if os.IsNotExist(err) {
			logger.Info("Destination file not exists. Sending original...")
			caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", m.Document.FileName, m.Sender.Username)
			videofile.Upload(bot, m, caption, UploadFinishedCallback)
		} else {
			logger.Info("Sending destination file...")
			caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>Original size: %.2f MB (%d kb/s)\nConverted size: %.2f MB (%d kb/s)</i>", m.Document.FileName, m.Sender.Username, float32(m.Document.FileSize)/1048576, srcBitrate/1024, float32(fi.Size())/1048576, dstBitrate/1024)
			videofile.filepath = dstPath // It's ok, cause we sending original
			videofile.Upload(bot, m, caption, UploadFinishedCallback)
		}
	} else {
		logger.Errorf("%s is not mpeg video", m.Document.MIME)
	}
}
