package usecases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

// Stat represents game statistics
type Stat struct {
	Player *core.User
	Score  int
}

// Find player by username in current stat
func Find(a []Stat, username string) int {
	for i, n := range a {
		if username == n.Player.Username {
			return i
		}
	}
	return -1
}
