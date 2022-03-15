package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateReelsFlow(l core.ILogger, mediaFactory core.IMediaFactory, sendMediaStrategy core.ISendMediaStrategy) core.ITextHandler {
	return &ReelsFlow{l, mediaFactory, sendMediaStrategy}
}

type ReelsFlow struct {
	l                 core.ILogger
	mediaFactory      core.IMediaFactory
	sendMediaStrategy core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *ReelsFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https://www.instagram.com/reel/\S+`)
	match := r.FindAllStringSubmatch(message.Text, -1)

	if len(match) < 1 {
		return fmt.Errorf("not implemented")
	}

	media, err := flow.mediaFactory.CreateMedia(match[0][0])
	if err != nil {
		flow.l.Error(err)
		return err
	}

	if len(media) < 1 {
		return fmt.Errorf("unexpected count of media")
	}

	m := &core.Media{
		ResourceURL: media[0].ResourceURL,
		Caption:     fmt.Sprintf("<a href='%s'>📷</a> <b>%s</b> <i>(by %s)</i>\n%s", match[0][0], media[0].Title, message.Sender.DisplayName(), media[0].Caption),
	}

	return flow.sendMediaStrategy.SendMedia([]*core.Media{m}, bot)
}
