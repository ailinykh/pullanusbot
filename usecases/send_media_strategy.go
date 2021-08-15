package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSendMediaStrategy(l core.ILogger) *SendMediaStrategy {
	return &SendMediaStrategy{l}
}

type SendMediaStrategy struct {
	l core.ILogger
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (sms *SendMediaStrategy) SendMedia(media []*core.Media, bot core.IBot) error {
	switch len(media) {
	case 0:
		sms.l.Warning("Unexpected empty media")
	case 1:
		_, err := bot.SendMedia(media[0])
		return err
	default:
		_, err := bot.SendPhotoAlbum(media)
		return err
	}
	return nil
}
