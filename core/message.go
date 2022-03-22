package core

// Message from chat
type Message struct {
	ID        int
	Chat      *Chat
	IsPrivate bool
	Sender    *User
	Text      string
	ReplyTo   *Message
	Video     *Video
}
