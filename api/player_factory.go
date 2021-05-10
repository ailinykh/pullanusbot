package api

import (
	"github.com/ailinykh/pullanusbot/v2/infrastructure"
	tb "gopkg.in/tucnak/telebot.v2"
)

type PlayerFactory struct {
	m *tb.Message
}

func (p *PlayerFactory) Make(string) infrastructure.Player {
	return infrastructure.Player{
		GameID:       p.m.Chat.ID,
		UserID:       p.m.Sender.ID,
		FirstName:    p.m.Sender.FirstName,
		LastName:     p.m.Sender.LastName,
		Username:     p.m.Sender.Username,
		LanguageCode: p.m.Sender.LanguageCode,
	}
}

// func makePlayer(m *tb.Message) infrastructure.Player {
// 	return infrastructure.Player{
// 		GameID:       m.Chat.ID,
// 		UserID:       m.Sender.ID,
// 		FirstName:    m.Sender.FirstName,
// 		LastName:     m.Sender.LastName,
// 		Username:     m.Sender.Username,
// 		LanguageCode: m.Sender.LanguageCode,
// 	}
// }
