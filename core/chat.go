package core

type ChatID = int64

type Chat struct {
	ID    ChatID
	Title string
	Type  string
}
