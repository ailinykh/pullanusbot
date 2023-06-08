package usecases

import (
	"fmt"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateSettingsFlow(l core.ILogger, loc core.ILocalizer, storage core.IBoolSettingProvider) *SettingsFlow {
	rootMenu := &SettingsMenu{
		id: "settings_root",
	}

	ids := []string{
		"instagram",
		"link",
		"tiktok",
		"twitter",
		"youtube",
	}

	for i, id := range ids {
		menu := &SettingsMenu{
			id: id,
			items: []SettingsItem{
				{
					id:      "settings_button_action",
					title:   loc.I18n("settings_flow"),
					setting: core.SettingKey(id + "_flow"),
				},
				{
					id:      "settings_button_action_2",
					title:   loc.I18n("settings_flow_remove_source"),
					setting: core.SettingKey(id + "_flow_remove_source"),
				},
				{
					id:    "settings_button_back",
					title: loc.I18n("settings_button_back"),
					menu:  rootMenu,
				},
			},
		}
		rootMenu.items = append(rootMenu.items, SettingsItem{
			id:    fmt.Sprintf("settings_button_forward_%d", i), // must be unique
			title: loc.I18n("settings_" + id),
			menu:  menu,
		})
	}

	rootMenu.items = append(rootMenu.items, SettingsItem{
		id:    "settings_button_cancel",
		title: loc.I18n("settings_button_cancel"),
	})

	return &SettingsFlow{l, loc, storage,
		rootMenu,
		make(map[core.ChatID]*SettingsMenu),
	}
}

type SettingsFlow struct {
	l       core.ILogger
	loc     core.ILocalizer
	storage core.IBoolSettingProvider
	menu    *SettingsMenu
	state   map[core.ChatID]*SettingsMenu
}

type SettingsMenu struct {
	id    string
	items []SettingsItem
}

type SettingsItem struct {
	id      string
	title   string
	setting core.SettingKey
	menu    *SettingsMenu
}

// GetButtonIds is a core.IButtonHandler protocol implementation
func (flow *SettingsFlow) GetButtonIds() []string {
	keys := []string{
		"settings_button_back",
		"settings_button_forward",
		"settings_button_cancel",
		"settings_button_enable",
		"settings_button_disable",
		"settings_button_action",
		"settings_button_action_2",
	}

	for i := range flow.menu.items {
		keys = append(keys, fmt.Sprintf("settings_button_forward_%d", i))
	}

	return keys
}

// ButtonPressed is a core.IButtonHandler protocol implementation
func (flow *SettingsFlow) ButtonPressed(button *core.Button, message *core.Message, user *core.User, bot core.IBot) error {
	flow.l.Infof("button pressed: %+v", button)
	if button.ID == "settings_button_cancel" {
		delete(flow.state, message.Chat.ID)
		_, err := bot.Edit(message, flow.loc.I18n("settings_canceled"))
		return err
	}

	if button.ID == "settings_button_back" {
		flow.state[message.Chat.ID] = flow.menu
		keyboard := flow.makeCurrentSettingsKeyboard(message.Chat.ID, flow.menu)
		_, err := bot.Edit(message, flow.loc.I18n("settings_title"), keyboard)
		return err
	}

	if button.ID == "settings_button_action" {
		// setting := core.SettingKey(button.Payload)
		// s := flow.storage.GetBool(message.Chat.ID, setting)
		// err := flow.storage.SetBool(message.Chat.ID, setting, !s)

	}

	if len(button.Payload) > 0 {
		if strings.HasPrefix(button.ID, "settings_button_forward") {
			id := button.Payload
			menu := flow.getMenuByID(id, flow.menu)
			if menu == nil {
				return fmt.Errorf("unexpected menu id: %s", id)
			}

			flow.state[message.Chat.ID] = menu
			keyboard := flow.makeCurrentSettingsKeyboard(message.Chat.ID, menu)
			messages := []string{
				"<b>" + flow.loc.I18n("settings_title_"+id) + "</b>",
				"",
				flow.loc.I18n("settings_description_" + id),
			}
			_, err := bot.Edit(message, strings.Join(messages, "\n"), keyboard)
			return err
		} else if strings.HasPrefix(button.ID, "settings_button_action") {
			settingKey := core.SettingKey(button.Payload)
			setting := flow.storage.GetBool(message.Chat.ID, settingKey)
			_ = flow.storage.SetBool(message.Chat.ID, settingKey, !setting)
			menu := flow.state[message.Chat.ID]
			keyboard := flow.makeCurrentSettingsKeyboard(message.Chat.ID, menu)
			messages := []string{
				"<b>" + flow.loc.I18n("settings_title_"+menu.id) + "</b>",
				"",
				flow.loc.I18n("settings_description_" + menu.id),
			}
			_, err := bot.Edit(message, strings.Join(messages, "\n"), keyboard)
			return err
		}

		flow.l.Errorf("payload not handled: %s", button.Payload)
	}

	return fmt.Errorf(button.ID + " not implemented")
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *SettingsFlow) HandleText(message *core.Message, bot core.IBot) error {
	if message.Text != "/settings" {
		return fmt.Errorf("not implemented")
	}

	keyboard := flow.makeCurrentSettingsKeyboard(message.Chat.ID, flow.menu)
	for _, btn := range keyboard {
		flow.l.Infof("%+v", btn[0])
	}
	_, err := bot.SendText(flow.loc.I18n("settings_title"), keyboard)
	return err
}

func (flow *SettingsFlow) getMenuByID(id string, menu *SettingsMenu) *SettingsMenu {
	flow.l.Infof("searching menu by id %s - %s", id, menu.id)
	if menu.id == id {
		return menu
	}

	for _, item := range menu.items {
		if item.menu != nil {
			if item.menu.id == id {
				return item.menu
			}
		}
	}

	return nil
}

func (flow *SettingsFlow) makeButtonFor(chatID core.ChatID, key core.SettingKey) (*core.Button, string) {
	s := flow.storage.GetBool(chatID, key)
	action := "enable"
	state := "disabled"
	if s {
		action = "disable"
		state = "enabled"
	}
	return &core.Button{
		ID:   fmt.Sprintf("settings_%s", action),
		Text: flow.loc.I18n("settings_button_" + action),
	}, state
}

func (flow *SettingsFlow) makeCurrentSettingsKeyboard(chatID core.ChatID, menu *SettingsMenu) core.Keyboard {
	keyboard := core.Keyboard{}

	for i, item := range menu.items {
		if item.menu != nil {
			// submenu item
			keyboard = append(keyboard, []*core.Button{
				{
					ID:      item.id,
					Text:    item.title,
					Payload: item.menu.id,
				},
			})
		} else if len(item.setting) > 0 {
			// setting
			s := flow.storage.GetBool(chatID, item.setting)
			flow.l.Infof("%s - %s - %t", item.id, item.setting, s)
			state := "disabled"
			if s {
				state = "enabled"
			}
			payload := fmt.Sprintf("%s", item.setting)
			state = flow.loc.I18n(fmt.Sprintf("settings_%s", state))
			text := fmt.Sprintf("%s - %s", item.title, state)
			keyboard = append(keyboard, []*core.Button{
				{
					ID:      item.id,
					Text:    text,
					Payload: payload,
				},
			})
		} else {
			// cancel
			keyboard = append(keyboard, []*core.Button{
				{ID: item.id, Text: item.title},
			})
		}
		flow.l.Infof("%d %+v %d", i, keyboard[len(keyboard)-1][0], len(keyboard[len(keyboard)-1][0].Payload))
	}

	return keyboard
}
