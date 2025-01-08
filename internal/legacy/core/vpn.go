package core

type VpnKey struct {
	ID     string
	ChatID ChatID
	Title  string
	Key    string
}

type IVpnAPI interface {
	GetKeys(ChatID) ([]*VpnKey, error)
	CreateKey(string, ChatID, *User) (*VpnKey, error)
	DeleteKey(*VpnKey) error
}
