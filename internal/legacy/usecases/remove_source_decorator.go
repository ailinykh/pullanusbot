package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateRemoveSourceDecorator(l core.ILogger, decoratee core.ITextHandler, settingsKey core.SettingKey, settingProvider core.IBoolSettingProvider) *RemoveSourceDecorator {
	return &RemoveSourceDecorator{l, decoratee, settingsKey, settingProvider}
}

type RemoveSourceDecorator struct {
	l               core.ILogger
	decoratee       core.ITextHandler
	settingsKey     core.SettingKey
	settingProvider core.IBoolSettingProvider
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

	enabled := decorator.settingProvider.GetBool(message.Chat.ID, decorator.settingsKey)

	if enabled {
		decorator.l.Infof("removing chat %d message %d", message.Chat.ID, message.ID)
		return bot.Delete(message)
	}

	return nil
}
