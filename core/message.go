package core

type Message struct {
	ID        int
	ChatID    int64
	IsPrivate bool
	Sender    *User
	Text      string
}
