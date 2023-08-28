package usecases

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/api"
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateInstagramFlow(l core.ILogger, api api.InstAPI, createVideo core.IVideoFactory, sendMedia core.ISendMediaStrategy, sendVideo core.ISendVideoStrategy) core.ITextHandler {
	return &InstagramFlow{l, api, createVideo, sendMedia, sendVideo}
}

type InstagramFlow struct {
	l           core.ILogger
	api         api.InstAPI
	createVideo core.IVideoFactory
	sendMedia   core.ISendMediaStrategy
	sendVideo   core.ISendVideoStrategy
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
	reel, err := flow.api.GetReel(url)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	if len(reel.Items) < 1 {
		return fmt.Errorf("insufficient reel items")
	}

	item := reel.Items[0]

	caption := item.Caption.Text
	if info := item.ClipsMetadata.MusicInfo; info != nil {
		caption = fmt.Sprintf("\n🎶 <a href='%s'>%s - %s</a>\n\n%s", info.MusicAssetInfo.ProgressiveDownloadURL, info.MusicAssetInfo.DisplayArtist, info.MusicAssetInfo.Title, caption)
	}
	caption = fmt.Sprintf("<a href='%s'>📷</a> <b>%s</b> <i>(by %s)</i>\n%s", url, item.User.FullName, message.Sender.DisplayName(), caption)
	if len(caption) > 1024 {
		// strip by last space or line break if caption size limit exceeded
		index := strings.LastIndex(caption[:1024], " ")
		lineBreak := strings.LastIndex(caption[:1024], "\n")
		if lineBreak > index {
			index = lineBreak
		}
		caption = caption[:index]
	}

	if item.VideoDuration < 360 { // apparently 6 min file takes less than 50 MB
		return flow.sendAsMedia(item, caption, message, bot)
	}

	video, err := flow.createVideo.CreateVideo(item.VideoVersions[0].URL)
	if err != nil {
		flow.l.Error(err)
		return err
	}
	defer video.Dispose()

	return flow.sendVideo.SendVideo(video, caption, bot)
}

func (flow *InstagramFlow) sendAsMedia(item api.IgReelItem, caption string, message *core.Message, bot core.IBot) error {
	media := &core.Media{
		ResourceURL: item.VideoVersions[0].URL,
		URL:         "https://www.instagram.com/reel/" + item.Code + "/",
		Title:       item.User.FullName,
		Caption:     caption,
	}

	err := flow.sendMedia.SendMedia([]*core.Media{media}, bot)
	if err != nil {
		flow.l.Error(err)
	}
	return err
}
