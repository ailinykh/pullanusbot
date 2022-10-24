package usecases

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateYoutubeFlow(l core.ILogger, mediaFactory core.IMediaFactory, videoFactory core.IVideoFactory, sendStrategy core.ISendVideoStrategy) *YoutubeFlow {
	return &YoutubeFlow{l: l, mediaFactory: mediaFactory, videoFactory: videoFactory, sendStrategy: sendStrategy}
}

type YoutubeFlow struct {
	mutex        sync.Mutex
	l            core.ILogger
	mediaFactory core.IMediaFactory
	videoFactory core.IVideoFactory
	sendStrategy core.ISendVideoStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *YoutubeFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`youtu\.?be(\.com)?\/(watch\?v=)?([\w\-_]+)`)
	match := r.FindStringSubmatch(message.Text)
	if len(match) == 4 {
		err := flow.process(match[3], message, bot)
		if err != nil {
			return err
		}

		if !strings.Contains(message.Text, " ") {
			return nil
		}
	} else if strings.Contains(message.Text, "youtu") {
		for i, m := range match {
			flow.l.Info(i, " ", m)
		}
		return fmt.Errorf("possibble regexp mismatch: %s", message.Text)
	}
	return fmt.Errorf("not implemented")
}

func (flow *YoutubeFlow) process(id string, message *core.Message, bot core.IBot) error {
	flow.mutex.Lock()
	defer flow.mutex.Unlock()

	flow.l.Infof("processing %s", id)
	media, err := flow.mediaFactory.CreateMedia(id)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	if !message.IsPrivate && media[0].Duration > 900 {
		flow.l.Infof("skip video in group chat due to duration %d", media[0].Duration)
		return fmt.Errorf("skip video in group chat due to duration")
	}

	title := media[0].Title
	file, err := flow.videoFactory.CreateVideo(id)
	if err != nil {
		return err
	}
	defer file.Dispose()

	caption := fmt.Sprintf(`<a href="https://youtu.be/%s">🎞</a> <b>%s</b> <i>(by %s)</i>`, id, title, message.Sender.DisplayName())
	return flow.sendStrategy.SendVideo(file, caption, bot)
}
