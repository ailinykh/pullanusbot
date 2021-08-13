package usecases_test

import "github.com/ailinykh/pullanusbot/v2/core"

type FakeBot struct {
	sentMessages    []string
	removedMessages []string
}

func (FakeBot) SendImage(*core.Image, string) (*core.Message, error)  { return nil, nil }
func (FakeBot) SendAlbum([]*core.Image) ([]*core.Message, error)      { return nil, nil }
func (FakeBot) SendMedia(*core.Media) (*core.Message, error)          { return nil, nil }
func (FakeBot) SendPhotoAlbum([]*core.Media) ([]*core.Message, error) { return nil, nil }
func (FakeBot) SendVideo(*core.Video, string) (*core.Message, error)  { return nil, nil }

func (b *FakeBot) Delete(message *core.Message) error {
	b.removedMessages = append(b.removedMessages, message.Text)
	return nil
}

func (b *FakeBot) SendText(text string, args ...interface{}) (*core.Message, error) {
	b.sentMessages = append(b.sentMessages, text)
	return nil, nil
}
