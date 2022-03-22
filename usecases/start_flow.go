package usecases

import (
	"strings"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateStartFlow(l core.ILogger, loc core.ILocalizer, chatStorage core.IChatStorage, userStorage core.IUserStorage) core.ITextHandler {
	return &StartFlow{l, loc, chatStorage, userStorage, make(map[int64]bool), make(map[int]bool), make(map[int64]bool), sync.Mutex{}}
}

type StartFlow struct {
	l             core.ILogger
	loc           core.ILocalizer
	chatStorage   core.IChatStorage
	userStorage   core.IUserStorage
	settingsCache map[int64]bool
	usersCache    map[core.UserID]bool
	chatCache     map[int64]bool
	lock          sync.Mutex
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *StartFlow) HandleText(message *core.Message, bot core.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	err := flow.ensureUserExists(message, bot)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	err = flow.ensureChatExists(message, bot)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	if strings.HasPrefix(message.Text, "/start") {
		if len(message.Text) > 7 {
			payload := message.Text[7:]
			flow.handlePayload(payload, message.Chat.ID)
			if err != nil {
				flow.l.Error(err)
				//Do not return?
			}
		}
		_, err = bot.SendText(flow.loc.I18n("start_welcome"))
		return err
	}

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

func (flow *StartFlow) ensureChatExists(message *core.Message, bot core.IBot) error {
	if _, ok := flow.chatCache[message.Chat.ID]; !ok {
		flow.chatCache[message.Chat.ID] = true
		_, err := flow.chatStorage.GetChatByID(message.Chat.ID)
		if err != nil {
			if err.Error() == "record not found" {
				settings := core.DefaultSettings()
				return flow.chatStorage.CreateChat(message.Chat.ID, message.Chat.Title, message.Chat.Type, &settings)
			}
			flow.l.Error(err)
			return err
		}
	}
	return nil
}

func (flow *StartFlow) ensureUserExists(message *core.Message, bot core.IBot) error {
	if _, ok := flow.usersCache[message.Sender.ID]; !ok {
		flow.usersCache[message.Sender.ID] = true
		_, err := flow.userStorage.GetUserById(message.Sender.ID)
		if err != nil {
			if err.Error() == "record not found" {
				return flow.userStorage.CreateUser(message.Sender)
			}
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
