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
func Find(a []Stat, id int) int {
	for i, n := range a {
		if id == n.Player.ID {
			return i
		}
	}
	return -1
}
