package helpers

import (
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateSendVideoStrategy(l core.ILogger) core.ISendVideoStrategy {
	return &SendVideoStrategy{l}
}

type SendVideoStrategy struct {
	l core.ILogger
}

// SendMedia is a core.ISendVideoStrategy interface implementation
func (strategy *SendVideoStrategy) SendVideo(video *core.Video, caption string, bot core.IBot) error {
	_, err := bot.SendVideo(video, caption)

	if err != nil {
		strategy.l.Error(err)
	}

	return err
}
