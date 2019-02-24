package main

import (
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Info is just a simple information helper
type Info struct {
}

// initialize database and all nesessary command handlers
func (i *Info) initialize() {
	log.Println("Info: subscribing to bot events")

	bot.Handle("/proxy", i.proxy)

	log.Println("Info: successfully initialized")
}

func (i *Info) proxy(m *tb.Message) {
	bot.Send(m.Chat, "tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
}
