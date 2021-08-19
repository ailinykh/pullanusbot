package usecases

import (
	"strings"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateIDoNotCare() *IDoNotCare {
	return &IDoNotCare{}
}

type IDoNotCare struct{}

// HandleText is a core.ITextHandler protocol implementation
func (IDoNotCare) HandleText(message *core.Message, bot core.IBot) error {
	if strings.Contains(strings.ToLower(message.Text), "мне всё равно") {
		_, err := bot.SendText("https://coub.com/view/1ov5oi", false)
		return err
	}
	if strings.Contains(strings.ToLower(message.Text), "привет, андрей") {
		_, err := bot.SendVideo(&core.Video{ID: "BAACAgIAAxkBAAIziWEeZBqlM1_1n2AVaxedGFn3vS-sAAKgDwACSl7xSImLuE-s8DMbIAQ"}, "")
		return err
	}
	return nil
}
