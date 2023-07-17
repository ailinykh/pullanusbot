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
	r := regexp.MustCompile(`youtu\.?be(\.com)?(\/shorts)?(\/live)?\/(watch\?v=)?([\w\-_]+)`)

	links := r.FindAllStringSubmatch(message.Text, -1)
	// TODO: any limits?
	for i, l := range links {
		err := flow.process(l[5], fmt.Sprintf("%s [%d/%d]", l[0], i+1, len(links)), message, bot)
		if err != nil {
			flow.l.Error(err)
			return err
		}
	}

	if len(links) > 0 && !strings.Contains(message.Text, " ") {
		return nil
	}
	// TODO: in case of `nil` the original message will be deleted
	return fmt.Errorf("not implemented")
}

func (flow *YoutubeFlow) process(id string, match string, message *core.Message, bot core.IBot) error {
	flow.mutex.Lock()
	defer flow.mutex.Unlock()

	flow.l.Infof("processing %s from %s", id, match)
	id = "https://youtu.be/" + id // -e9_M7-0quU
	media, err := flow.mediaFactory.CreateMedia(id)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	flow.l.Infof("video: %s %.2f MB %d sec, audio: %s %.2f MB", media[0].Codec, float64(media[0].Size)/1024/1024, media[0].Duration, media[1].Codec, float64(media[1].Size)/1024/1024)

	totlalSize := media[0].Size + media[1].Size
	if !message.IsPrivate && totlalSize > 50_000_000 {
		flow.l.Infof("skip video in group chat due to size limit exceeded %d", totlalSize)
		return nil // TODO: should return error?
	}

	file, err := flow.videoFactory.CreateVideo(id)
	if err != nil {
		return err
	}
	defer file.Dispose()

	caption := fmt.Sprintf("<a href=\"%s\">🎞</a> <b>%s</b> <i>(by %s)</i>\n\n%s", id, media[0].Title, message.Sender.DisplayName(), media[0].Description)
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	caption = strings.ToValidUTF8(caption, "")
	return flow.sendStrategy.SendVideo(file, caption, bot)
}
