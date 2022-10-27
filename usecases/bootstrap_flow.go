package usecases

import (
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateBootstrapFlow(l core.ILogger, chatStorage core.IChatStorage, userStorage core.IUserStorage) core.ITextHandler {
	return &BootstrapFlow{l, chatStorage, userStorage, sync.Mutex{}}
}

type BootstrapFlow struct {
	l           core.ILogger
	chatStorage core.IChatStorage
	userStorage core.IUserStorage
	lock        sync.Mutex
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *BootstrapFlow) HandleText(message *core.Message, bot core.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	err := flow.ensureUserExists(message.Sender)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	err = flow.ensureChatExists(message.Chat)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	return err
}

func (flow *BootstrapFlow) ensureChatExists(chat *core.Chat) error {
	_, err := flow.chatStorage.GetChatByID(chat.ID)
	if err != nil {
		if err.Error() == "record not found" {
			return flow.chatStorage.CreateChat(chat.ID, chat.Title, chat.Type)
		}
		flow.l.Error(err)
	}
	return err
}

func (flow *BootstrapFlow) ensureUserExists(user *core.User) error {
	_, err := flow.userStorage.GetUserById(user.ID)
	if err != nil {
		if err.Error() == "record not found" {
			return flow.userStorage.CreateUser(user)
		}
		flow.l.Error(err)
	}
	return err
}
