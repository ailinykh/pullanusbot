package usecases_test

import (
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_NotFoundAnyLinkByDefault(t *testing.T) {
	handler := &FakeTweetHandler{[]string{}}
	parser := usecases.CreateTwitterParser(handler)
	m := makeTweetMessage("a message without any links")
	bot := &BotMock{}

	parser.HandleText(m, bot)

	assert.Equal(t, []string{}, handler.tweets)
}

func Test_HandleText_FoundTweetLink(t *testing.T) {
	handler := &FakeTweetHandler{[]string{}}
	parser := usecases.CreateTwitterParser(handler)
	m := makeTweetMessage("a message with https://twitter.com/status/username/123456")
	bot := &BotMock{}

	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456"}, handler.tweets)
}

func Test_HandleText_FoundMultipleTweetLinks(t *testing.T) {
	handler := &FakeTweetHandler{[]string{}}
	parser := usecases.CreateTwitterParser(handler)
	m := makeTweetMessage("a message with https://twitter.com/status/username/123456 and https://twitter.com/status/username/789010 and some text")
	bot := &BotMock{}

	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456", "789010"}, handler.tweets)
}

func makeTweetMessage(text string) *core.Message {
	return &core.Message{ID: 0, Text: text}
}

type FakeTweetHandler struct {
	tweets []string
}

// HandleTweet is a ITweetHandler protocol implementation
func (fth *FakeTweetHandler) HandleTweet(tweetID string, message *core.Message, bot core.IBot) error {
	fth.tweets = append(fth.tweets, tweetID)
	return nil
}
