package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateRemoveSourceDecorator(l core.ILogger, decoratee core.ITextHandler, settingsStorage core.ISettingsStorage) *RemoveSourceDecorator {
	return &RemoveSourceDecorator{l, decoratee, settingsStorage}
}

type RemoveSourceDecorator struct {
	l               core.ILogger
	decoratee       core.ITextHandler
	settingsStorage core.ISettingsStorage
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

	settings, err := decorator.settingsStorage.GetSettings(message.ChatID)
	if err != nil {
		decorator.l.Error(err)
		return err
	}

	if settings.RemoveSourceOnSucccess {
		decorator.l.Infof("removing chat %d message %d", message.ChatID, message.ID)
		return bot.Delete(message)
	}

	return nil
}
