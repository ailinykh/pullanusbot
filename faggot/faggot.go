package faggot

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Player struct for serialization
type Player struct {
	*tb.User
}

// Entry struct for game result serialization
type Entry struct {
	Day      string `json:"day"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func (p *Player) mention() string {
	var str strings.Builder
	str.WriteString("@")
	str.WriteString(p.Username)
	return str.String()
}
