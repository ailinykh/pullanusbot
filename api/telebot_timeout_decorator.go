package api

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateTelebotTimeoutDecorator(l core.ILogger, d core.IBot) *TelebotTimeoutDecorator {
	return &TelebotTimeoutDecorator{l, d}
}

type TelebotTimeoutDecorator struct {
	l core.ILogger
	d core.IBot
}

// SendText is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendText(text string, params ...interface{}) (*core.Message, error) {
	m, err := ttd.d.SendText(text, params...)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// Delete is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) Delete(message *core.Message) error {
	err := ttd.d.Delete(message)
	if err != nil {
		ttd.l.Error(err)
	}
	return err
}

// SendImage is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendImage(image *core.Image, caption string) (*core.Message, error) {
	m, err := ttd.d.SendImage(image, caption)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// SendAlbum is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendAlbum(images []*core.Image) ([]*core.Message, error) {
	m, err := ttd.d.SendAlbum(images)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// SendMedia is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendMedia(media *core.Media) (*core.Message, error) {
	m, err := ttd.d.SendMedia(media)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// SendPhotoAlbum is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendPhotoAlbum(medias []*core.Media) ([]*core.Message, error) {
	m, err := ttd.d.SendPhotoAlbum(medias)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// SendVideo is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) SendVideo(vf *core.Video, caption string) (*core.Message, error) {
	m, err := ttd.d.SendVideo(vf, caption)
	if err != nil {
		ttd.l.Error(err)
	}
	return m, err
}

// IsUserMemberOfChat is a core.IBot interface implementation
func (ttd *TelebotTimeoutDecorator) IsUserMemberOfChat(user *core.User, chatID int64) bool {
	return ttd.d.IsUserMemberOfChat(user, chatID)
}
