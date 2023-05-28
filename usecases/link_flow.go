package usecases

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// CreateLinkFlow is a basic LinkFlow factory
func CreateLinkFlow(l core.ILogger, httpClient core.IHttpClient, mediaFactory core.IMediaFactory, sendMediaStrategy core.ISendMediaStrategy) *LinkFlow {
	return &LinkFlow{l, httpClient, mediaFactory, sendMediaStrategy}
}

// LinkFlow converts hotlink to video/photo attachment
type LinkFlow struct {
	l                 core.ILogger
	httpClient        core.IHttpClient
	mediaFactory      core.IMediaFactory
	sendMediaStrategy core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *LinkFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(message.Text) {
		return flow.handleURL(message.Text, message, bot)
	}
	return fmt.Errorf("not implemented")
}

func (flow *LinkFlow) handleURL(url core.URL, message *core.Message, bot core.IBot) error {
	contentType, err := flow.httpClient.GetContentType(url)
	if err != nil {
		// skip "content-type not found"
		return nil
	}

	if !strings.HasPrefix(contentType, "video") && !strings.HasPrefix(contentType, "image") {
		return fmt.Errorf("not implemented")
	}

	media, err := flow.mediaFactory.CreateMedia(url)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	for _, m := range media {
		switch m.Type {
		case core.TPhoto:
			m.Caption = fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.DisplayName())
		case core.TVideo:
			m.Caption = fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.DisplayName())
		case core.TText:
			flow.l.Warningf("Unexpected %+v", m)
		}
	}

	err = flow.sendMediaStrategy.SendMedia(media, bot)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	return nil
}
