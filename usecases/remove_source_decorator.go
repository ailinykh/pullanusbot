package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateRemoveSourceDecorator(l core.ILogger, decoratee core.ITextHandler) *RemoveSourceDecorator {
	return &RemoveSourceDecorator{l, decoratee}
}

type RemoveSourceDecorator struct {
	l         core.ILogger
	decoratee core.ITextHandler
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

	if message.Chat.Settings.RemoveSourceOnSucccess {
		decorator.l.Infof("removing chat %d message %d", message.Chat.ID, message.ID)
		return bot.Delete(message)
	}

	return nil
}
