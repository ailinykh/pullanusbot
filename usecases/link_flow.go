package usecases

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateLinkFlow is a basic LinkFlow factory
func CreateLinkFlow(l core.ILogger, hc core.IHttpClient, mf core.IMediaFactory, ms core.ISendMediaStrategy) *LinkFlow {
	return &LinkFlow{l, hc, mf, ms}
}

// LinkFlow converts hotlink to video/photo attachment
type LinkFlow struct {
	l   core.ILogger
	hc  core.IHttpClient
	mf  core.IMediaFactory
	sms core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (lf *LinkFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(message.Text) {
		return lf.handleURL(message, bot)
	}
	return nil
}

func (lf *LinkFlow) handleURL(message *core.Message, bot core.IBot) error {
	contentType, err := lf.hc.GetContentType(message.Text)
	if err != nil {
		lf.l.Error(err)
		return err
	}

	if !strings.HasPrefix(contentType, "video") && !strings.HasPrefix(contentType, "image") {
		return nil
	}

	media, err := lf.mf.CreateMedia(message.Text)
	if err != nil {
		lf.l.Error(err)
		return err
	}

	for _, m := range media {
		switch m.Type {
		case core.TPhoto:
			m.Caption = fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.Username)
		case core.TVideo:
			m.Caption = fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.Username)
		case core.TText:
			lf.l.Warningf("Unexpected %+v", m)
		}
	}

	err = lf.sms.SendMedia(media, bot)
	if err != nil {
		lf.l.Error(err)
		return err
	}

	return bot.Delete(message)
}
