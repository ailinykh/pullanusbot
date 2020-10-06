package interfaces

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// Bot is a generic interface for mocks
type Bot interface {
	ChatMemberOf(*tb.Chat, *tb.User) (*tb.ChatMember, error)
	Delete(tb.Editable) error
	Download(*tb.File, string) error
	Handle(interface{}, interface{})
	Notify(tb.Recipient, tb.ChatAction) error
	Send(tb.Recipient, interface{}, ...interface{}) (*tb.Message, error)
	SendAlbum(tb.Recipient, tb.Album, ...interface{}) ([]tb.Message, error)
	Start()
}

// IBotAdapter is a generic interface for different bot communication structs
type IBotAdapter interface {
	Setup(Bot, *gorm.DB)
}

// TextMessageHandler is interface to receive `text` messages, not only `commands`
type TextMessageHandler interface {
	HandleTextMessage(*tb.Message)
}
