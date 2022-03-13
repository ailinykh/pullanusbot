package test_helpers

import (
	"fmt"

	"github.com/ailinykh/pullanusbot/v2/core"
)

// https://stackoverflow.com/questions/31794141/can-i-create-shared-test-utilities

func CreateBot() *FakeBot {
	return &FakeBot{[]string{}, []string{}, []string{}, []string{}, map[int64][]string{}}
}

type FakeBot struct {
	SentMedias      []string
	SentMessages    []string
	SentVideos      []string
	RemovedMessages []string
	ChatMembers     map[int64][]string
}

func (FakeBot) SendImage(*core.Image, string) (*core.Message, error) { return nil, nil }
func (FakeBot) SendAlbum([]*core.Image) ([]*core.Message, error)     { return nil, nil }

func (b *FakeBot) SendMedia(media *core.Media) (*core.Message, error) {
	b.SentMedias = append(b.SentMedias, media.ResourceURL)
	return nil, nil
}

func (b *FakeBot) SendPhotoAlbum(media []*core.Media) ([]*core.Message, error) {
	for _, m := range media {
		b.SentMedias = append(b.SentMedias, m.ResourceURL)
	}
	return nil, nil
}

func (b *FakeBot) SendVideo(video *core.Video, caption string) (*core.Message, error) {
	b.SentVideos = append(b.SentVideos, video.Name)
	return nil, nil
}

func (b *FakeBot) Delete(message *core.Message) error {
	b.RemovedMessages = append(b.RemovedMessages, message.Text)
	return nil
}

func (b *FakeBot) Edit(message *core.Message, what interface{}, options ...interface{}) (*core.Message, error) {
	return nil, fmt.Errorf("not implemented")
}

func (b *FakeBot) SendText(text string, args ...interface{}) (*core.Message, error) {
	b.SentMessages = append(b.SentMessages, text)
	return nil, nil
}

func (b *FakeBot) IsUserMemberOfChat(user *core.User, chatID int64) bool {
	for _, username := range b.ChatMembers[chatID] {
		if username == user.Username {
			return true
		}
	}
	return false
}
