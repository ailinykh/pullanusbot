package usecases

import (
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterParser(tt *TwitterTimeout) *TwitterParser {
	return &TwitterParser{tt}
}

type TwitterParser struct {
	tt *TwitterTimeout
}

// HandleText is a core.ITextHandler protocol implementation
func (tp *TwitterParser) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`twitter\.com.+/(\d+)\S*$`)
	match := r.FindStringSubmatch(message.Text)
	if len(match) < 2 {
		return nil // no tweet id found
	}
	return tp.tt.HandleTweet(match[1], message, bot)
}
