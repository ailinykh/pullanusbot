package usecases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateStartFlow(l core.ILogger, loc core.ILocalizer, settingsStorage core.ISettingsStorage) core.ITextHandler {
	return &StartFlow{l, loc, settingsStorage}
}

type StartFlow struct {
	l               core.ILogger
	loc             core.ILocalizer
	settingsStorage core.ISettingsStorage
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *StartFlow) HandleText(message *core.Message, bot core.IBot) error {
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

func (flow *StartFlow) contains(payload string, current []string) bool {
	for _, p := range current {
		if p == payload {
			return true
		}
	}
	return false
}
