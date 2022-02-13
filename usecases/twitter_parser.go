package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/pullanusbot/v2/core"
)

func CreateTwitterParser(l core.ILogger, th ITweetHandler) *TwitterParser {
	return &TwitterParser{l, th}
}

type TwitterParser struct {
	l  core.ILogger
	th ITweetHandler
}

// HandleText is a core.ITextHandler protocol implementation
func (tp *TwitterParser) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https://twitter\.com\S+/(\d+)\S*`)
	match := r.FindAllStringSubmatch(message.Text, -1)

	if len(match) > 0 {
		tp.l.Infof("Processing %s", match[0][0])
	} else {
		return fmt.Errorf("not implemented")
	}

	for _, m := range match {
		err := tp.th.HandleTweet(m[1], message, bot)
		if err != nil {
			return err
		}
	}

	return nil
}
