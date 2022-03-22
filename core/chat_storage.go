package core

type IChatStorage interface {
	GetChatByID(int64) (*Chat, error)
	CreateChat(int64, string, string, *Settings) error
}
