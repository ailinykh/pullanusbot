package main

import (
	"log"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	logfile, err := os.OpenFile("data/log.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Println("BOT_TOKEN not set")
		return
	}

	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: poller,
	})

	if err != nil {
		log.Println(err)
		return
	}

	game := NewFaggotGame(bot)
	game.Start()

	bot.Start()
}
