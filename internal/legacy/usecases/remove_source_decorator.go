package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateRemoveSourceDecorator(l core.Logger, decoratee legacy.ITextHandler, settingsKey legacy.SettingKey, settingProvider legacy.IBoolSettingProvider) *RemoveSourceDecorator {
	return &RemoveSourceDecorator{l, decoratee, settingsKey, settingProvider}
}

type RemoveSourceDecorator struct {
	l               core.Logger
	decoratee       legacy.ITextHandler
	settingsKey     legacy.SettingKey
	settingProvider legacy.IBoolSettingProvider
}

// HandleText is a core.ITextHandler protocol implementation
func (decorator *RemoveSourceDecorator) HandleText(message *legacy.Message, bot legacy.IBot) error {
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
		decorator.l.Info("removing chat %d message %d", message.Chat.ID, message.ID)
		return bot.Delete(message)
	}

	return nil
}
