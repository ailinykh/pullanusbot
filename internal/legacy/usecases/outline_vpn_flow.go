package usecases

import (
	"fmt"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateVpnFlow(l core.ILogger, loc core.ILocalizer, api core.IVpnAPI) core.ITextHandler {
	flow := OutlineVpnFlow{l, loc, api, make(map[string]func(*core.Message, core.IBot) error), make(map[int64]OutlineVpnState)}
	flow.callbacks["vpn_create_key"] = flow.create
	flow.callbacks["vpn_manage_key"] = flow.manage
	flow.callbacks["vpn_delete_key"] = flow.delete
	flow.callbacks["vpn_back"] = flow.back
	flow.callbacks["vpn_cancel"] = flow.cancel
	return &flow
}

type OutlineVpnFlow struct {
	l   core.ILogger
	loc core.ILocalizer
	api core.IVpnAPI

	callbacks map[string]func(*core.Message, core.IBot) error
	state     map[int64]OutlineVpnState
}

type OutlineVpnState struct {
	action string
	source int
}

// GetButtonIds is a core.IButtonHandler protocol implementation
func (flow *OutlineVpnFlow) GetButtonIds() []string {
	keys := make([]string, len(flow.callbacks))

	i := 0
	for k := range flow.callbacks {
		keys[i] = k
		i++
	}

	return keys
}

// ButtonPressed is a core.IButtonHandler protocol implementation
func (flow *OutlineVpnFlow) ButtonPressed(button *core.Button, message *core.Message, _ *core.User, bot core.IBot) error {
	if callback, ok := flow.callbacks[button.ID]; ok {
		return callback(message, bot)
	}
	return fmt.Errorf("not implemented")
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *OutlineVpnFlow) HandleText(message *core.Message, bot core.IBot) error {
	if !message.IsPrivate {
		return fmt.Errorf("not implemented")
	}

	if state, ok := flow.state[message.Chat.ID]; ok {
		return flow.handleAction(state, message, bot)
	}

	if message.Text != "/vpnhelp" {
		return fmt.Errorf("not implemented")
	}

	return flow.help(message, bot)
}

func (flow *OutlineVpnFlow) help(message *core.Message, bot core.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_welcome"), flow.getKeyboard(message, keys))
	return err
}

func (flow *OutlineVpnFlow) create(message *core.Message, bot core.IBot) error {
	flow.state[message.Chat.ID] = OutlineVpnState{"create", message.ID}
	keyboard := core.Keyboard{[]*core.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}}}

	_, err := bot.Edit(message, flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_create_key_name"), keyboard)
	return err
}

func (flow *OutlineVpnFlow) manage(message *core.Message, bot core.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	text := []string{flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_top")}

	for idx, key := range keys {
		text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_item", idx+1, key.Title, key.Key))
	}

	text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_bottom", len(keys)))

	keyboard := core.Keyboard{
		[]*core.Button{{ID: "vpn_delete_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_remove_key")}},
		[]*core.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}},
	}
	_, err = bot.Edit(message, strings.Join(text, "\n"), keyboard)
	return err
}

func (flow *OutlineVpnFlow) back(message *core.Message, bot core.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	delete(flow.state, message.Chat.ID)

	_, err = bot.Edit(message, flow.loc.I18n(message.Sender.LanguageCode, "vpn_welcome"), flow.getKeyboard(message, keys))
	return err
}

func (flow *OutlineVpnFlow) delete(message *core.Message, bot core.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	flow.state[message.Chat.ID] = OutlineVpnState{"delete", message.ID}

	text := []string{flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_delete_key_name_top")}

	for _, key := range keys {
		text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_delete_key_name_item", key.Title))
	}

	keyboard := core.Keyboard{[]*core.Button{
		{ID: "vpn_cancel", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_cancel")},
	}}
	_, err = bot.Edit(message, strings.Join(text, "\n"), keyboard)
	return err
}

func (flow *OutlineVpnFlow) cancel(message *core.Message, bot core.IBot) error {
	return flow.back(message, bot)
}

func (flow *OutlineVpnFlow) getKeyboard(message *core.Message, keys []*core.VpnKey) core.Keyboard {
	keyboard := core.Keyboard{}

	if len(keys) < 10 {
		keyboard = append(keyboard, []*core.Button{{ID: "vpn_create_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_create_key")}})
	}

	if len(keys) > 0 {
		keyboard = append(keyboard, []*core.Button{{ID: "vpn_manage_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_manage_key")}})
	}

	return keyboard
}

func (flow *OutlineVpnFlow) handleAction(state OutlineVpnState, message *core.Message, bot core.IBot) error {
	if state.action == "create" {
		if len(message.Text) > 64 {
			_, err := bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_create_key_name_too_long"))
			return err
		}
		key, err := flow.api.CreateKey(message.Chat.ID, message.Text)
		if err != nil {
			flow.l.Error(err)
			return err
		}

		delete(flow.state, message.Chat.ID)

		_ = bot.Delete(&core.Message{ID: state.source, Chat: message.Chat})

		keyboard := core.Keyboard{[]*core.Button{{ID: "vpn_manage_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_manage_key")}}}
		_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_created", key.Key), keyboard)
		return err
	}

	if state.action == "delete" {
		keys, err := flow.api.GetKeys(message.Chat.ID)
		if err != nil {
			flow.l.Error(err)
			return err
		}

		delete(flow.state, message.Chat.ID)

		_ = bot.Delete(&core.Message{ID: state.source, Chat: message.Chat})

		keyboard := core.Keyboard{
			[]*core.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}},
		}

		for _, k := range keys {
			if k.Title == message.Text {
				err = flow.api.DeleteKey(k)
				if err != nil {
					return err
				}
				_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_deleted", k.Title), keyboard)
				return err
			}
		}
		_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_not_found"), keyboard)
		return err
	}

	return fmt.Errorf("unexpected action: %s", state.action)
}
