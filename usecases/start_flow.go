package usecases

import (
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateStartFlow(l core.ILogger, loc core.ILocalizer, settingsStorage core.ISettingsStorage) core.ITextHandler {
	return &StartFlow{l, loc, settingsStorage, make(map[int64]bool)}
}

type StartFlow struct {
	l               core.ILogger
	loc             core.ILocalizer
	settingsStorage core.ISettingsStorage
	cache           map[int64]bool
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *StartFlow) HandleText(message *core.Message, bot core.IBot) error {
	if strings.HasPrefix(message.Text, "/start") {
		return flow.handleStart(message, bot)
	}
	return flow.handleSettingsChack(message, bot)
}

func (flow *StartFlow) handleStart(message *core.Message, bot core.IBot) error {
	settings, err := flow.settingsStorage.GetSettings(message.ChatID)
	if err != nil {
		return err
	}

	if len(message.Text) > 7 {
		payload := message.Text[7:]
		if !flow.contains(payload, settings.Payload) {
			settings.Payload = append(settings.Payload, payload)
			err = flow.settingsStorage.SetSettings(message.ChatID, settings)
			if err != nil {
				return err
			}
		}
	}
	_, err = bot.SendText(flow.loc.I18n("start_welcome"))
	return err
}

func (flow *StartFlow) handleSettingsChack(message *core.Message, bot core.IBot) error {
	if _, ok := flow.cache[message.ChatID]; !ok {
		flow.cache[message.ChatID] = true
		_, err := flow.settingsStorage.GetSettings(message.ChatID) // create settings if needed
		if err != nil {
			flow.l.Error(err)
			return err
		}
	}
	return nil
}

func (flow *StartFlow) contains(payload string, current []string) bool {
	for _, p := range current {
		if p == payload {
			return true
		}
	}
	return false
}
