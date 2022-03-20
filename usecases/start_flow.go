package usecases

import (
	"strings"
	"sync"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateStartFlow(l core.ILogger, loc core.ILocalizer, settingsStorage core.ISettingsStorage, userStorage core.IUserStorage) core.ITextHandler {
	return &StartFlow{l, loc, settingsStorage, userStorage, make(map[int64]bool), make(map[int]bool), sync.Mutex{}}
}

type StartFlow struct {
	l               core.ILogger
	loc             core.ILocalizer
	settingsStorage core.ISettingsStorage
	userStorage     core.IUserStorage
	settingsCache   map[int64]bool
	usersCache      map[core.UserID]bool
	lock            sync.Mutex
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *StartFlow) HandleText(message *core.Message, bot core.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	if strings.HasPrefix(message.Text, "/start") {
		return flow.handleStart(message, bot)
	}
	return flow.checkSettingsAndUserPresence(message, bot)
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

func (flow *StartFlow) checkSettingsAndUserPresence(message *core.Message, bot core.IBot) error {
	if _, ok := flow.settingsCache[message.ChatID]; !ok {
		flow.l.Infof("%+v %+v", message, message.Sender)
		flow.settingsCache[message.ChatID] = true
		_, err := flow.settingsStorage.GetSettings(message.ChatID) // create settings if needed
		if err != nil {
			flow.l.Error(err)
			return err
		}
	}

	if _, ok := flow.usersCache[message.Sender.ID]; !ok {
		flow.l.Infof("%+v %+v", message, message.Sender)
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
