package api

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Telebot struct {
	bot *tb.Bot
}

func CreateTelebot(token string) *Telebot {
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

	return &Telebot{bot}
}

func (t *Telebot) Run() {
	t.bot.Start()
}
