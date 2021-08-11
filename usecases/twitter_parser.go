package usecases

import (
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterParser(th ITweetHandler) *TwitterParser {
	return &TwitterParser{th}
}

type TwitterParser struct {
	th ITweetHandler
}

// HandleText is a core.ITextHandler protocol implementation
func (tp *TwitterParser) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https://twitter\.com\S+/(\d+)\S*`)
	match := r.FindAllStringSubmatch(message.Text, -1)

	for _, m := range match {
		err := tp.th.HandleTweet(m[1], message, bot, message.Text == m[0])
		if err != nil {
			return err
		}
	}

	return nil
}
