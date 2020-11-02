package info

import (
	"fmt"
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
	bot.Handle("/info", i.info)
}

func (i *Info) proxy(m *tb.Message) {
	bot.Send(m.Chat, "tg://proxy?server=proxy.ailinykh.com&port=443&secret=dd71ce3b5bf1b7015dc62a76dc244c5aec")
}

func (i *Info) info(m *tb.Message) {
	info := fmt.Sprintf("💬 Chat\nID: *%d*\nTitle: *%s*\nType: *%s*\n\n👤 User\nID: *%d*\nFirst: *%s*\nLast: *%s*\n", m.Chat.ID, m.Chat.Title, m.Chat.Type, m.Sender.ID, m.Sender.FirstName, m.Sender.LastName)
	bot.Send(m.Chat, info, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}
