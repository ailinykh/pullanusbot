package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

type ITweetHandler interface {
	HandleTweet(string, *core.Message, core.IBot, bool) error
}

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterFlow(l core.ILogger, mf core.IMediaFactory, sms core.ISendMediaStrategy) *TwitterFlow {
	return &TwitterFlow{l, mf, sms}
}

// TwitterFlow represents tweet processing logic
type TwitterFlow struct {
	l   core.ILogger
	mf  core.IMediaFactory
	sms core.ISendMediaStrategy
}

// HandleTweet is a ITweetHandler protocol implementation
func (tf *TwitterFlow) HandleTweet(tweetID string, message *core.Message, bot core.IBot, deleteOriginal bool) error {
	tf.l.Infof("processing tweet %s, delete original: %t", tweetID, deleteOriginal)
	media, err := tf.mf.CreateMedia(tweetID, message.Sender)
	if err != nil {
		tf.l.Error(err)
		return err
	}

	err = tf.sms.SendMedia(media, bot)
	if err != nil {
		tf.l.Error(err)
		return err
	}

	if deleteOriginal {
		return bot.Delete(message)
	}

	return nil
}
