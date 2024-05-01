package usecases

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/api"
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateInstagramFlow(l core.ILogger, api api.YoutubeApi, sendMedia core.ISendMediaStrategy) core.ITextHandler {
	return &InstagramFlow{l, api, sendMedia}
}

type InstagramFlow struct {
	l         core.ILogger
	api       api.YoutubeApi
	sendMedia core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *InstagramFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https://www.instagram.com/reel/\S+`)
	rmatch := r.FindAllString(message.Text, -1)

	switch len(rmatch) {
	case 0:
		break
	case 1:
		return flow.handleReel(rmatch[0], message, bot)
	default:
		for _, reel := range rmatch {
			err := flow.handleReel(reel, message, bot)
			if err != nil {
				flow.l.Error(err)
				return err
			}
		}
		// FIXME: temporal coupling
		return fmt.Errorf("do not remove source message")
	}

	t := regexp.MustCompile(`https://www.instagram.com/tv/\S+`)
	tmatch := t.FindAllString(message.Text, -1)

	// TODO: multiple tv?
	if len(tmatch) > 0 {
		return flow.handleReel(tmatch[0], message, bot)
	}

	return fmt.Errorf("not implemented")
}

func (flow *InstagramFlow) handleReel(url string, message *core.Message, bot core.IBot) error {
	flow.l.Infof("processing %s", url)
	resp, err := flow.api.Get(url)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	caption := fmt.Sprintf("<a href='%s'>📷</a> <b>%s</b> <i>(by %s)</i>\n%s", url, resp.Uploader, message.Sender.DisplayName(), resp.Description)
	if len(caption) > 1024 {
		// strip by last space or line break if caption size limit exceeded
		index := strings.LastIndex(caption[:1024], " ")
		lineBreak := strings.LastIndex(caption[:1024], "\n")
		if lineBreak > index {
			index = lineBreak
		}
		caption = caption[:index]
	}

	vf, err := flow.getPreferredVideoFormat(resp)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	media := core.Media{
		Caption:     caption,
		ResourceURL: vf.Url,
		URL:         url,
	}
	return flow.sendMedia.SendMedia([]*core.Media{&media}, bot)
}

func (flow *InstagramFlow) getPreferredVideoFormat(resp *api.YtDlpResponse) (*api.YtDlpFormat, error) {
	idx := -1
	for i, f := range resp.Formats {
		if strings.HasPrefix(f.FormatId, "dash-") {
			continue
		}
		idx = i
	}

	if idx < 0 {
		return nil, fmt.Errorf("no appropriate format found")
	}
	return resp.Formats[idx], nil
}
