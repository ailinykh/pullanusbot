package usecases_test

import (
	"testing"

	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/ailinykh/pullanusbot/v2/test_helpers"
	"github.com/ailinykh/pullanusbot/v2/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_CreatesSettingsWithPayload(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	settingsStorage := test_helpers.CreateSettingsStorage()
	userStorage := test_helpers.CreateUserStorage()
	startFlow := usecases.CreateStartFlow(logger, loc, settingsStorage, userStorage)

	message := &core.Message{Text: "/start payload", ChatID: 1488}
	bot := test_helpers.CreateBot()

	startFlow.HandleText(message, bot)

	expectedSettings := core.DefaultSettings()
	expectedSettings.Payload = []string{"payload"}
	assert.Equal(t, map[int64]*core.Settings{1488: &expectedSettings}, settingsStorage.Data)
}

func Test_HandleText_MergePayloadInSettings(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	settingsStorage := test_helpers.CreateSettingsStorage()
	userStorage := test_helpers.CreateUserStorage()
	startFlow := usecases.CreateStartFlow(logger, loc, settingsStorage, userStorage)

	initialSettings := core.DefaultSettings()
	initialSettings.Payload = []string{"payload"}
	settingsStorage.SetSettings(1488, &initialSettings)
	bot := test_helpers.CreateBot()

	startFlow.HandleText(&core.Message{Text: "/start another_payload", ChatID: 1488}, bot)

	expectedSettings := core.DefaultSettings()
	expectedSettings.Payload = []string{"payload", "another_payload"}
	assert.Equal(t, map[int64]*core.Settings{1488: &expectedSettings}, settingsStorage.Data)

	startFlow.HandleText(&core.Message{Text: "/start payload", ChatID: 1488}, bot)
	assert.Equal(t, map[int64]*core.Settings{1488: &expectedSettings}, settingsStorage.Data)
}
