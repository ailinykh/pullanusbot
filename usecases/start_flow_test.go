package usecases_test

import (
	"sync"
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/test_helpers"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_CreateChatPayload(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	chatStorage := test_helpers.CreateChatStorage()
	commandService := test_helpers.CreateCommandService(logger)
	startFlow := usecases.CreateStartFlow(logger, loc, chatStorage, commandService)

	settings := core.DefaultSettings()
	chatStorage.CreateChat(1488, "Paul Durov", "private", &settings)
	bot := test_helpers.CreateBot()

	messages := []string{
		"/start",
		"/start payload",
		"/start another_payload",
	}
	wg := sync.WaitGroup{}

	for _, message := range messages {
		wg.Add(1)
		go func(text string) {
			startFlow.Start(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(chatStorage.Chats))

	message := makeMessage("/start")
	chat, _ := chatStorage.GetChatByID(message.Chat.ID)
	assert.Equal(t, true, contains("payload", chat.Settings.Payload))
	assert.Equal(t, true, contains("another_payload", chat.Settings.Payload))

	expected := []string{
		"enable commands 1488 [{help show help message}]",
		"enable commands 1488 [{help show help message}]",
		"enable commands 1488 [{help show help message}]",
	}
	assert.Equal(t, expected, commandService.ActionLog)
}

func makeMessage(text string) *core.Message {
	settings := core.DefaultSettings()
	chat := core.Chat{ID: 1488, Title: "Paul Durov", Type: "private", Settings: &settings}
	sender := core.User{ID: 1, FirstName: "Paul", LastName: "Durov"}
	return &core.Message{Text: text, Chat: &chat, Sender: &sender}
}

func contains(message string, messages []string) bool {
	for _, m := range messages {
		if m == message {
			return true
		}
	}
	return false
}
