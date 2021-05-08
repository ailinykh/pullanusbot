package api

import (
	"time"

	"github.com/ailinykh/pullanusbot/v2/core"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Telebot struct {
	bot              *tb.Bot
	logger           core.ILogger
	videoFileFactory core.IVideoFileFactory
}

func CreateTelebot(token string, logger core.ILogger, videoFileFactory core.IVideoFileFactory) *Telebot {
	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	var err error
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		panic(err)
	}

	return &Telebot{bot, logger, videoFileFactory}
}

func (t *Telebot) Run() {
	t.bot.Start()
}
