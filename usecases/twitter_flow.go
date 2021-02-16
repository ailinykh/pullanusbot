package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

type ITweetHandler interface {
	Process(string, *core.Message, core.IBot) error
}

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterFlow(l core.ILogger, mediaFactory core.IMediaFactory, sendMediaStrategy core.ISendMediaStrategy) *TwitterFlow {
	return &TwitterFlow{l, mediaFactory, sendMediaStrategy}
}

// TwitterFlow represents tweet processing logic
type TwitterFlow struct {
	l                 core.ILogger
	mediaFactory      core.IMediaFactory
	sendMediaStrategy core.ISendMediaStrategy
}

// Process is a ITweetHandler protocol implementation
func (flow *TwitterFlow) Process(tweetID string, message *core.Message, bot core.IBot) error {
	flow.l.Infof("processing tweet %s", tweetID)
	media, err := flow.mediaFactory.CreateMedia(tweetID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	for _, m := range media {
		re := regexp.MustCompile(`\s?http\S+$`)
		text := re.ReplaceAllString(m.Description, "")
		m.Caption = fmt.Sprintf("<a href='%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", m.URL, m.Title, message.Sender.DisplayName(), text)
	}

	err = flow.sendMediaStrategy.SendMedia(media, bot)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	return nil
}
