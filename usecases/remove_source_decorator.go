package usecases

import (
	"encoding/json"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateRemoveSourceDecorator(l core.ILogger, decoratee core.ITextHandler, settingsKey core.SettingKey, settingsProvider core.ISettingsProvider) *RemoveSourceDecorator {
	return &RemoveSourceDecorator{l, decoratee, settingsKey, settingsProvider}
}

type RemoveSourceDecorator struct {
	l                core.ILogger
	decoratee        core.ITextHandler
	settingsKey      core.SettingKey
	settingsProvider core.ISettingsProvider
}

// HandleText is a core.ITextHandler protocol implementation
func (decorator *RemoveSourceDecorator) HandleText(message *core.Message, bot core.IBot) error {
	err := decorator.decoratee.HandleText(message, bot)
	//TODO: error handling protocol
	if err != nil && err.Error() == "not implemented" {
		return nil
	}

	if err != nil {
		decorator.l.Error(err)
		return err
	}

	data, _ := decorator.settingsProvider.GetData(message.Chat.ID, decorator.settingsKey)

	var settingsV1 struct {
		Enabled bool
	}

	err = json.Unmarshal(data, &settingsV1)
	if err != nil {
		decorator.l.Error(err)
		// TODO: perform a migration
		return nil
	}

	if settingsV1.Enabled {
		decorator.l.Infof("removing chat %d message %d", message.Chat.ID, message.ID)
		return bot.Delete(message)
	}

	return nil
}
