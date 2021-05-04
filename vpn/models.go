package vpn

import (
	"time"

	"github.com/google/logger"
)

// VpnKey represents vpn key
type VpnKey struct {
	UserID   int    `gorm:"primaryKey"`
	Date     string `gorm:"primaryKey"`
	Device   string
	Username string
	KeyData  []byte
}

// TableName for VPN keys
func (VpnKey) TableName() string {
	return "vpn_keys"
}

func (key *VpnKey) timestamp() time.Time {
	ts, err := time.Parse(time.RFC3339, key.Date)

	if err != nil {
		logger.Error(err)
	}

	return ts
}

func (key *VpnKey) unixTime() int64 {
	return key.timestamp().Unix()
}

func (key *VpnKey) humanReadableTime() string {
	return key.timestamp().Format("02 Jan 2006 15:04")
}
