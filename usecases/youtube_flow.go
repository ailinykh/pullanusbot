package usecases

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYoutubeFlow(l core.ILogger, mf core.IMediaFactory, vff core.IVideoFileFactory, vfs core.IVideoFileSplitter) *YoutubeFlow {
	return &YoutubeFlow{l: l, mf: mf, vff: vff, vfs: vfs}
}

type YoutubeFlow struct {
	m   sync.Mutex
	l   core.ILogger
	mf  core.IMediaFactory
	vff core.IVideoFileFactory
	vfs core.IVideoFileSplitter
}

// HandleText is a core.ITextHandler protocol implementation
func (f *YoutubeFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https?:\/\/(www\.)?youtu[\.be|\.com]\S+`)
	match := r.FindStringSubmatch(message.Text)
	if len(match) > 0 {
		return f.process(match[0], message, bot)
	}
	return nil
}

func (f *YoutubeFlow) process(url string, message *core.Message, bot core.IBot) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.l.Infof("processing youtube %s", url)
	media, err := f.mf.CreateMedia(url, message.Sender)
	if err != nil {
		return err
	}

	if !message.IsPrivate && media[0].Duration > 900 {
		f.l.Infof("skip video in group chat due to duration %d", media[0].Duration)
		return nil
	}

	youtubeID, title := media[0].URL, media[0].Caption
	f.l.Infof("downloading %s", youtubeID)
	file, err := f.vff.CreateVideoFile(youtubeID)
	if err != nil {
		return err
	}
	defer file.Dispose()

	caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, youtubeID, title, message.Sender.Username)
	_, err = bot.SendVideoFile(file, caption)
	if err != nil {
		f.l.Error("Can't send video: ", err)
		if err.Error() == "telegram: Request Entity Too Large (400)" {
			f.l.Info("Fallback to splitting")
			files, err := f.vfs.Split(file, 50000000)
			if err != nil {
				return err
			}

			for _, file := range files {
				defer file.Dispose()
			}

			for i, file := range files {
				caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🔗</a> <b>[%d/%d] %s</b> <i>(by %s)</i>`, youtubeID, i+1, len(files), title, message.Sender.Username)
				_, err := bot.SendVideoFile(file, caption)
				if err != nil {
					return err
				}
			}

			f.l.Info("All parts successfully sent")
			return bot.Delete(message)
		}
		return err
	}
	return bot.Delete(message)
}