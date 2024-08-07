package core

type IChatStorage interface {
	GetChatByID(ChatID) (*Chat, error)
	CreateChat(ChatID, string, string) error
}
