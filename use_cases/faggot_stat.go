package use_cases

import (
	"github.com/ailinykh/pullanusbot/v2/core"
)

type Stat struct {
	Player *core.User
	Score  int
}

func Find(a []Stat, username string) int {
	for i, n := range a {
		if username == n.Player.Username {
			return i
		}
	}
	return -1
}
