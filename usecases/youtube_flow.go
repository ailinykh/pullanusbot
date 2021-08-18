package usecases

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYoutubeFlow(l core.ILogger, mf core.IMediaFactory, vff core.IVideoFactory, vfs core.IVideoSplitter) *YoutubeFlow {
	return &YoutubeFlow{l: l, mf: mf, vff: vff, vfs: vfs}
}

type YoutubeFlow struct {
	m   sync.Mutex
	l   core.ILogger
	mf  core.IMediaFactory
	vff core.IVideoFactory
	vfs core.IVideoSplitter
}

// HandleText is a core.ITextHandler protocol implementation
func (f *YoutubeFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`youtu\.?be(\.com)?\/(watch\?v=)?([\w\-_]+)`)
	match := r.FindStringSubmatch(message.Text)
	if len(match) == 4 {
		return f.process(match[3], message, bot)
	}
	if strings.Contains(message.Text, "youtu") {
		for i, m := range match {
			f.l.Info(i, " ", m)
		}
		return errors.New("possibble regexp mismatch: " + message.Text)
	}
	return nil
}

func (f *YoutubeFlow) process(id string, message *core.Message, bot core.IBot) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.l.Infof("processing %s", id)
	media, err := f.mf.CreateMedia(id, message.Sender)
	if err != nil {
		return err
	}

	if !message.IsPrivate && media[0].Duration > 900 {
		f.l.Infof("skip video in group chat due to duration %d", media[0].Duration)
		return nil
	}

	title := media[0].Title
	f.l.Infof("downloading %s", id)
	file, err := f.vff.CreateVideo(id)
	if err != nil {
		return err
	}
	defer file.Dispose()

	caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, id, title, message.Sender.Username)
	_, err = bot.SendVideo(file, caption)
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
				caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>[%d/%d] %s</b> <i>(by %s)</i>`, id, i+1, len(files), title, message.Sender.Username)
				_, err := bot.SendVideo(file, caption)
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
