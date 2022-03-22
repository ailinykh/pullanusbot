package usecases_test

import (
	"sync"
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/test_helpers"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_CreateUserData(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	settingsStorage := test_helpers.CreateSettingsStorage()
	chatStorage := test_helpers.CreateChatStorage()
	userStorage := test_helpers.CreateUserStorage()
	startFlow := usecases.CreateStartFlow(logger, loc, settingsStorage, chatStorage, userStorage)

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
			startFlow.HandleText(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(userStorage.Users))

	message := makeMessage("/start")
	user, _ := userStorage.GetUserById(message.Sender.ID)
	assert.Equal(t, message.Sender, user)
}

func Test_HandleText_CreateChatData(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	settingsStorage := test_helpers.CreateSettingsStorage()
	chatStorage := test_helpers.CreateChatStorage()
	userStorage := test_helpers.CreateUserStorage()
	startFlow := usecases.CreateStartFlow(logger, loc, settingsStorage, chatStorage, userStorage)

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
			startFlow.HandleText(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(chatStorage.Chats))

	message := makeMessage("/start")
	chat, _ := chatStorage.GetChatByID(message.Chat.ID)
	assert.Equal(t, true, contains("payload", chat.Settings.Payload))
	assert.Equal(t, true, contains("another_payload", chat.Settings.Payload))
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
