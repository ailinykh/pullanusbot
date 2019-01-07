package main

import (
	"log"
	"os"
	"time"

	"./faggot"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("BOT_TOKEN not set")
	}

	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		log.Fatal(err)
	}

	faggot.Setup(bot)

	bot.Start()
}
