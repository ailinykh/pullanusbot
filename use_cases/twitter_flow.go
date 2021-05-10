package use_cases

import "github.com/ailinykh/pullanusbot/v2/core"

func CreateTwitterFlow(l core.ILogger) *TwitterFlow {
	return &TwitterFlow{l}
}

type TwitterFlow struct {
	l core.ILogger
}

func (tf *TwitterFlow) HandleText(text string, bot core.IBot) error {
	tf.l.Infof("Got message %s", text)
	return nil
}
