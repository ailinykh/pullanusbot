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
	return nil
}
