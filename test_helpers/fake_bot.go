package test_helpers

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateFakeBot() *FakeBot {
	return &FakeBot{[]string{}, []string{}, []string{}}
}

type FakeBot struct {
	SentMedias      []string
	SentMessages    []string
	RemovedMessages []string
}

func (FakeBot) SendImage(*core.Image, string) (*core.Message, error) { return nil, nil }
func (FakeBot) SendAlbum([]*core.Image) ([]*core.Message, error)     { return nil, nil }

func (b *FakeBot) SendMedia(media *core.Media) (*core.Message, error) {
	b.SentMedias = append(b.SentMedias, media.URL)
	return nil, nil
}

func (b *FakeBot) SendPhotoAlbum(media []*core.Media) ([]*core.Message, error) {
	for _, m := range media {
		b.SentMedias = append(b.SentMedias, m.URL)
	}
	return nil, nil
}

func (FakeBot) SendVideo(*core.Video, string) (*core.Message, error) { return nil, nil }

func (b *FakeBot) Delete(message *core.Message) error {
	b.RemovedMessages = append(b.RemovedMessages, message.Text)
	return nil
}

func (b *FakeBot) SendText(text string, args ...interface{}) (*core.Message, error) {
	b.SentMessages = append(b.SentMessages, text)
	return nil, nil
}
