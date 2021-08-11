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
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message with https://twitter.com/status/username/123456")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456"}, handler.tweets)
}

func Test_HandleText_FoundMultipleTweetLinks(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message with https://twitter.com/username/status/123456 and https://twitter.com/username/status/789010 and some text")
	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456", "789010"}, handler.tweets)
}

func Test_HandleText_RemovesOriginalMessageInCaseOfFullMatch(t *testing.T) {
	parser, _, bot := makeTwitterSUT()
	m := makeTweetMessage("https://twitter.com/username/status/123456")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{"https://twitter.com/username/status/123456"}, bot.removedMessages)
}

func Test_HandleText_DoesNotRemoveOriginalMessage(t *testing.T) {
	parser, _, bot := makeTwitterSUT()
	m := makeTweetMessage("https://twitter.com/username/status/123456 and some other text")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{}, bot.removedMessages)
}

func makeTwitterSUT() (*usecases.TwitterParser, *FakeTweetHandler, *BotMock) {
	handler := &FakeTweetHandler{[]string{}}
	parser := usecases.CreateTwitterParser(handler)
	bot := &BotMock{[]string{}, []string{}}
	return parser, handler, bot
}

func makeTweetMessage(text string) *core.Message {
	return &core.Message{ID: 0, Text: text}
}

type FakeTweetHandler struct {
	tweets []string
}

// HandleTweet is a ITweetHandler protocol implementation
func (fth *FakeTweetHandler) HandleTweet(tweetID string, message *core.Message, bot core.IBot, deleteOriginal bool) error {
	fth.tweets = append(fth.tweets, tweetID)
	if deleteOriginal {
		return bot.Delete(message)
	}
	return nil
}
