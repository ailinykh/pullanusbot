package usecases

import (
	"errors"

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
		if media[0].Type == core.TVideo && media[0].Codec != "mp4" {
			return errors.New("unexpected video codec " + media[0].Codec)
		}
		_, err := bot.SendMedia(media[0])
		return err
	default:
		_, err := bot.SendPhotoAlbum(media)
		return err
	}
	return nil
}
