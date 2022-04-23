package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateInstagramFlow(l core.ILogger, mediaFactory core.IMediaFactory, sendMedia core.ISendMediaStrategy, sendVideo core.ISendVideoStrategy) core.ITextHandler {
	return &InstagramFlow{l, mediaFactory, sendMedia, sendVideo}
}

type InstagramFlow struct {
	l            core.ILogger
	mediaFactory core.IMediaFactory
	sendMedia    core.ISendMediaStrategy
	sendVideo    core.ISendVideoStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *InstagramFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https://www.instagram.com/reel/\S+`)
	match := r.FindAllStringSubmatch(message.Text, -1)

	if len(match) > 0 {
		return flow.handleReel(match[0][0], message, bot)
	}

	return fmt.Errorf("not implemented")
}

func (flow *InstagramFlow) handleReel(url string, message *core.Message, bot core.IBot) error {
	media, err := flow.mediaFactory.CreateMedia(url)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	if len(media) < 1 {
		return fmt.Errorf("unexpected count of media")
	}

	m := &core.Media{
		ResourceURL: media[0].ResourceURL,
		Caption:     fmt.Sprintf("<a href='%s'>📷</a> <b>%s</b> <i>(by %s)</i>\n%s", url, media[0].Title, message.Sender.DisplayName(), media[0].Caption),
	}

	return flow.sendMedia.SendMedia([]*core.Media{m}, bot)
}
