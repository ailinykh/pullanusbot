package faggot

import (
	"strings"
	"sync"

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

// ConcurrentSlice is a thread safe integer store
type ConcurrentSlice struct {
	sync.RWMutex
	items []int64
}

// Add item to slice
func (s *ConcurrentSlice) Add(e int64) {
	s.Lock()
	defer s.Unlock()

	s.items = append(s.items, e)
}

// Remove item from slice
func (s *ConcurrentSlice) Remove(e int64) {
	s.Lock()
	defer s.Unlock()

	i := s.Index(e)
	s.items = append(s.items[0:i], s.items[i+1:]...)
}

// Index of current item or -1 if not found
func (s *ConcurrentSlice) Index(e int64) int {
	for i, el := range s.items {
		if el == e {
			return i
		}
	}
	return -1
}
