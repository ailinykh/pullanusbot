package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/logger"
	"github.com/stretchr/testify/assert"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	messages []tb.Message
)

type Bot struct{}

func (Bot) ChatMemberOf(c *tb.Chat, u *tb.User) (*tb.ChatMember, error)        { return nil, nil }
func (Bot) Delete(tb.Editable) error                                           { return nil }
func (Bot) Download(*tb.File, string) error                                    { return nil }
func (Bot) Edit(tb.Editable, interface{}, ...interface{}) (*tb.Message, error) { return nil, nil }
func (Bot) Handle(interface{}, interface{})                                    {}
func (Bot) Notify(tb.Recipient, tb.ChatAction) error                           { return nil }
func (Bot) Respond(*tb.Callback, ...*tb.CallbackResponse) error                { return nil }
func (Bot) Start()                                                             {}

func (Bot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	var m tb.Message
	switch object := what.(type) {
	case string:
		m = tb.Message{Text: object}
	case *tb.Photo:
		m = tb.Message{Photo: object}
	case *tb.Video:
		m = tb.Message{Video: object}
	default:
		logger.Errorf("%#v", object)
		return nil, tb.ErrUnsupportedWhat
	}
	messages = append(messages, m)
	return &m, nil
}

func (Bot) SendAlbum(m tb.Recipient, album tb.Album, params ...interface{}) ([]tb.Message, error) {
	for _, file := range album {
		m := tb.Message{Text: file.MediaFile().FileID}
		messages = append(messages, m)
	}
	return nil, nil
}

type HelperMock struct {
	h Helper
}

func (h HelperMock) getTweet(tweetID string) (Tweet, error) {
	file, _ := ioutil.ReadFile(fmt.Sprintf("fixtures/%s.json", tweetID))
	var tweet Tweet
	_ = json.Unmarshal([]byte(file), &tweet)
	return tweet, nil
}

func (h HelperMock) makeAlbum(media []Media, caption string) tb.Album {
	return h.h.makeAlbum(media, caption)
}

func (h HelperMock) makeCaption(m *tb.Message, tweet Tweet) string {
	return h.h.makeCaption(m, tweet)
}

func tearUp(t *testing.T) {
	logger.Init("twitter_tests", false, false, ioutil.Discard)
	helper = HelperMock{h: Helper{}}
	bot = Bot{}
	messages = []tb.Message{}
}

func TestRateLimitErrorInvokesGetTweetAfterTimeout(t *testing.T) {
	tearUp(t)
	tweet := Tweet{ID: "single_text", Errors: []Error{{Code: 88}}, Header: http.Header{"X-Rate-Limit-Reset": []string{"3"}}}
	tw := Twitter{}
	tw.processError(tweet, &tb.Message{Sender: &tb.User{Username: ""}})
}
func TestSendsTextAsText(t *testing.T) {
	tearUp(t)
	tw := Twitter{}
	tw.processTweet("single_text", &tb.Message{Sender: &tb.User{Username: ""}})
	assert.Contains(t, messages[0].Text, "Happy Birthday, Go!")
	assert.Nil(t, messages[0].Photo)
	assert.Nil(t, messages[0].Video)
}

func TestSendsSingleImageAsImageWithCaption(t *testing.T) {
	tearUp(t)
	tw := Twitter{}
	tw.processTweet("image_and_text", &tb.Message{Sender: &tb.User{Username: ""}})
	assert.Empty(t, messages[0].Text)
	assert.Contains(t, messages[0].Photo.Caption, "dog had my")
	assert.Nil(t, messages[0].Video)
}

func TestSendsMultipleImagesAsAlbumCommand(t *testing.T) {
	tearUp(t)
	tw := Twitter{}
	tw.processTweet("multiple_images", &tb.Message{Sender: &tb.User{Username: ""}})
	assert.Len(t, messages, 3)
}

func TestSendsGIFAsVideoWithCaption(t *testing.T) {
	tearUp(t)
	tw := Twitter{}
	tw.processTweet("gif_and_text", &tb.Message{Sender: &tb.User{Username: ""}})
	assert.Empty(t, messages[0].Text)
	assert.Contains(t, messages[0].Video.Caption, "Telegram")
	assert.Nil(t, messages[0].Photo)
}

func TestSendsVideoAsVideoWithCaption(t *testing.T) {
	tearUp(t)
	tw := Twitter{}
	tw.processTweet("video_and_text", &tb.Message{Sender: &tb.User{Username: ""}})
	assert.Empty(t, messages[0].Text)
	assert.Contains(t, messages[0].Video.Caption, "Space_Station in low-Earth orbit this year")
	assert.Nil(t, messages[0].Photo)
}
