package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTikTokFlow(l core.ILogger, hc core.IHttpClient, mf core.IMediaFactory, sms core.ISendMediaStrategy) *TikTokFlow {
	return &TikTokFlow{l, hc, mf, sms}
}

type TikTokFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	mf  core.IMediaFactory
	sms core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (ttf *TikTokFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https?://\w+\.tiktok.com/\S+`)
	links := r.FindAllString(message.Text, -1)
	for _, l := range links {
		err := ttf.handleURL(l, message, bot)
		if err != nil {
			return err
		}
	}

	if len(links) > 0 {
		return bot.Delete(message)
	}
	return nil
}

func (ttf *TikTokFlow) handleURL(url string, message *core.Message, bot core.IBot) error {
	ttf.l.Info("processing ", url)
	HTMLString, err := ttf.hc.GetContent(url)
	if err != nil {
		return err
	}

	media, err := ttf.mf.CreateMedia(HTMLString)
	if err != nil {
		return err
	}

	for _, m := range media {
		m.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", m.URL, m.Title, message.Sender.Username, m.Description)
	}
	return ttf.sms.SendMedia(media, bot)
}
