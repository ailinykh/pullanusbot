package usecases_test

import "github.com/ailinykh/pullanusbot/v2/core"

type BotMock struct {
	sentMessages    []string
	removedMessages []string
}

func (BotMock) SendImage(*core.Image, string) (*core.Message, error)  { return nil, nil }
func (BotMock) SendAlbum([]*core.Image) ([]*core.Message, error)      { return nil, nil }
func (BotMock) SendMedia(*core.Media) (*core.Message, error)          { return nil, nil }
func (BotMock) SendPhotoAlbum([]*core.Media) ([]*core.Message, error) { return nil, nil }
func (BotMock) SendVideo(*core.Video, string) (*core.Message, error)  { return nil, nil }

func (b *BotMock) Delete(message *core.Message) error {
	b.removedMessages = append(b.removedMessages, message.Text)
	return nil
}

func (b *BotMock) SendText(text string, args ...interface{}) (*core.Message, error) {
	b.sentMessages = append(b.sentMessages, text)
	return nil, nil
}
