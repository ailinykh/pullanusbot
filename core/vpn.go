package core

type VpnKey struct {
	ID     string
	ChatID int64
	Title  string
	Key    string
}

type IVpnAPI interface {
	GetKeys(int64) ([]*VpnKey, error)
	CreateKey(int64, string) (*VpnKey, error)
	DeleteKey(*VpnKey) error
}
