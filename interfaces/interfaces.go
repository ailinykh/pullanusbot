package interfaces

import tb "gopkg.in/tucnak/telebot.v2"

type Bot interface {
	ChatMemberOf(*tb.Chat, *tb.User) (*tb.ChatMember, error)
	Handle(interface{}, interface{})
	Send(tb.Recipient, interface{}, ...interface{}) (*tb.Message, error)
}
