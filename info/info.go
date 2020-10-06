package info

import (
	i "pullanusbot/interfaces"

	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

var (
	bot i.Bot
)

// Info is just a simple information helper
type Info struct {
}

// Setup all nesessary command handlers
func (i *Info) Setup(b i.Bot, conn *gorm.DB) {
	bot = b
	bot.Handle("/proxy", i.proxy)
}

func (i *Info) proxy(m *tb.Message) {
	bot.Send(m.Chat, "tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
}
