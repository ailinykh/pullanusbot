package info

import (
	"fmt"
	i "pullanusbot/interfaces"
	"strings"

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
	info := []string{
		"💬 Chat",
		fmt.Sprintf("ID: *%d*", m.Chat.ID),
		fmt.Sprintf("Title: *%s*", m.Chat.Title),
		fmt.Sprintf("Type: *%s*", m.Chat.Type),
		"",
		"👤 Sender",
		fmt.Sprintf("ID: *%d*", m.Sender.ID),
		fmt.Sprintf("First: *%s*", m.Sender.FirstName),
		fmt.Sprintf("Last: *%s*", m.Sender.LastName),
		"",
	}

	if m.ReplyTo != nil {
		if m.ReplyTo.OriginalChat != nil {
			info = append(info,
				"💬 OriginalChat",
				fmt.Sprintf("ID: *%d*", m.ReplyTo.OriginalChat.ID),
				fmt.Sprintf("Title: *%s*", m.ReplyTo.OriginalChat.Title),
				fmt.Sprintf("Type: *%s*", m.ReplyTo.OriginalChat.Type),
				"",
			)
		}
		if m.ReplyTo.OriginalSender != nil {
			info = append(info,
				"👤 OriginalSender",
				fmt.Sprintf("ID: *%d*", m.ReplyTo.OriginalSender.ID),
				fmt.Sprintf("First: *%s*", m.ReplyTo.OriginalSender.FirstName),
				fmt.Sprintf("Last: *%s*", m.ReplyTo.OriginalSender.LastName),
				"",
			)
		}
	}

	bot.Send(m.Chat, strings.Join(info, "\n"), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}
