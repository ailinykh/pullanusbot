package usecases

import (
	"strings"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateStartFlow(l core.ILogger, loc core.ILocalizer, chatStorage core.IChatStorage, commandService core.ICommandService) *StartFlow {
	return &StartFlow{l, loc, chatStorage, commandService, sync.Mutex{}}
}

type StartFlow struct {
	l              core.ILogger
	loc            core.ILocalizer
	chatStorage    core.IChatStorage
	commandService core.ICommandService
	lock           sync.Mutex
}

func (flow *StartFlow) Start(message *core.Message, bot core.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	if strings.HasPrefix(message.Text, "/start") {
		if len(message.Text) > 7 {
			payload := message.Text[7:]
			err := flow.handlePayload(payload, message.Chat.ID)
			if err != nil {
				flow.l.Error(err)
				//return err ?
			}
		}

		err := flow.commandService.EnableCommands(message.Chat.ID, []core.Command{{Text: "help", Description: "show help message"}}, bot)
		if err != nil {
			flow.l.Error(err)
			// return err ?
		}
		_, err = bot.SendText(flow.loc.I18n("start_welcome") + " " + flow.loc.I18n("help"))
		return err
	}

	return nil
}

func (flow *StartFlow) Help(message *core.Message, bot core.IBot) error {
	_, err := bot.SendText(flow.loc.I18n("help"))
	return err
}

func (flow *StartFlow) handlePayload(payload string, chatID int64) error {
	chat, err := flow.chatStorage.GetChatByID(chatID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	if flow.contains(payload, chat.Settings.Payload) {
		return nil
	}

	chat.Settings.Payload = append(chat.Settings.Payload, payload)
	return flow.chatStorage.UpdateSettings(chat.ID, chat.Settings)
}

func (flow *StartFlow) contains(payload string, current []string) bool {
	for _, p := range current {
		if p == payload {
			return true
		}
	}
	return false
}
