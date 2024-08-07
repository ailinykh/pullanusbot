package api

import (
	"fmt"
	"strings"

	tb "gopkg.in/telebot.v3"
)

// SetupInfo ...
func (t *Telebot) SetupInfo() {
	t.bot.Handle("/info", func(c tb.Context) error {
		m := c.Message()
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
			fmt.Sprintf("Username: *%s*", m.Sender.Username),
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
					fmt.Sprintf("Username: *%s*", m.ReplyTo.OriginalSender.Username),
					fmt.Sprintf("Name: *%s*", m.ReplyTo.OriginalSenderName),
					"",
				)
			}
		}

		_, err := t.bot.Send(m.Chat, strings.Join(info, "\n"), &tb.SendOptions{ParseMode: tb.ModeMarkdown})
		return err
	})
}
