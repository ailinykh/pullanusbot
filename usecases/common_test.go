package usecases_test

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateFakeBot() *FakeBot {
	return &FakeBot{[]string{}, []string{}, []string{}}
}

type FakeBot struct {
	sentMedias      []string
	sentMessages    []string
	removedMessages []string
}

func (FakeBot) SendImage(*core.Image, string) (*core.Message, error) { return nil, nil }
func (FakeBot) SendAlbum([]*core.Image) ([]*core.Message, error)     { return nil, nil }

func (b *FakeBot) SendMedia(media *core.Media) (*core.Message, error) {
	b.sentMedias = append(b.sentMedias, media.URL)
	return nil, nil
}

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
