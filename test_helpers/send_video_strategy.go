package test_helpers

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSendVideoStrategy() *FakeSendVideoStrategy {
	return &FakeSendVideoStrategy{[]string{}, nil}
}

type FakeSendVideoStrategy struct {
	SentVideos []string
	Err        error
}

// SendVideo is a core.ISendVideoStrategy interface implementation
func (fsms *FakeSendVideoStrategy) SendVideo(video *core.Video, caption string, bot core.IBot) error {
	fsms.SentVideos = append(fsms.SentVideos, video.Name)
	return fsms.Err
}
