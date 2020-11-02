package converter

import (
	"os"
	"path"
	"pullanusbot/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	tb "gopkg.in/tucnak/telebot.v2"
)

const ShortVideoURL = "https://www.w3schools.com/html/mov_bbb.mp4"

var (
	messages  []*tb.Message
	converter *Converter
	filePath  string
	user1     *tb.User
	user2     *tb.User
	chat      *tb.Chat
)

type Bot struct{}

func (Bot) ChatMemberOf(c *tb.Chat, u *tb.User) (*tb.ChatMember, error)            { return nil, nil }
func (Bot) Delete(tb.Editable) error                                               { return nil }
func (Bot) Download(*tb.File, string) error                                        { return nil }
func (Bot) Edit(tb.Editable, interface{}, ...interface{}) (*tb.Message, error)     { return nil, nil }
func (Bot) Handle(interface{}, interface{})                                        {}
func (Bot) Notify(tb.Recipient, tb.ChatAction) error                               { return nil }
func (Bot) Respond(*tb.Callback, ...*tb.CallbackResponse) error                    { return nil }
func (Bot) SendAlbum(tb.Recipient, tb.Album, ...interface{}) ([]tb.Message, error) { return nil, nil }
func (Bot) Start()                                                                 {}

func (Bot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	m := &tb.Message{Text: what.(string)}
	messages = append(messages, m)
	return m, nil
}

func tearUp(t *testing.T) func() {
	converter = &Converter{}
	converter.Setup(&Bot{}, nil)
	messages = []*tb.Message{}

	if filePath == "" {
		filePath = path.Join(os.TempDir(), "tmp-video-file"+utils.RandStringRunes(4)+"mp4")
	}
	return func() {
	}
}

func TestConverterIgnoresAudio(t *testing.T) {
	defer tearUp(t)()
	m := &tb.Message{Document: &tb.Document{MIME: "audio/mp3"}}
	converter.checkMessage(m)
	assert.Equal(t, len(messages), 0)
}
